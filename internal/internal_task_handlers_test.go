// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/zap"
)

const (
	testDomain = "test-domain"
)

type (
	TaskHandlersTestSuite struct {
		suite.Suite
		logger  *zap.Logger
		service *workflowservicetest.MockClient
	}
)

func init() {
	RegisterWorkflowWithOptions(
		helloWorldWorkflowFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_Workflow"},
	)
	RegisterWorkflowWithOptions(
		helloWorldWorkflowCancelFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_WorkflowCancel"},
	)
	RegisterWorkflowWithOptions(
		returnPanicWorkflowFunc,
		RegisterWorkflowOptions{Name: "ReturnPanicWorkflow"},
	)
	RegisterWorkflowWithOptions(
		panicWorkflowFunc,
		RegisterWorkflowOptions{Name: "PanicWorkflow"},
	)
	RegisterWorkflowWithOptions(
		getWorkflowInfoWorkflowFunc,
		RegisterWorkflowOptions{Name: "GetWorkflowInfoWorkflow"},
	)
	RegisterActivityWithOptions(
		greeterActivityFunc,
		RegisterActivityOptions{Name: "Greeter_Activity"},
	)
}

func returnPanicWorkflowFunc(ctx Context, input []byte) error {
	return newPanicError("panicError", "stackTrace")
}

func panicWorkflowFunc(ctx Context, input []byte) error {
	panic("panicError")
}

func getWorkflowInfoWorkflowFunc(ctx Context, expectedLastCompletionResult string) (info *WorkflowInfo, err error) {
	result := GetWorkflowInfo(ctx)
	var lastCompletionResult string
	err = getDefaultDataConverter().FromData(result.lastCompletionResult, &lastCompletionResult)
	if err != nil {
		return nil, err
	}
	if lastCompletionResult != expectedLastCompletionResult {
		return nil, errors.New("lastCompletionResult is not " + expectedLastCompletionResult)
	}
	return result, nil
}

// Test suite.
func (t *TaskHandlersTestSuite) SetupTest() {
}

func (t *TaskHandlersTestSuite) SetupSuite() {
	logger, _ := zap.NewDevelopment()
	t.logger = logger
}

func TestTaskHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(TaskHandlersTestSuite))
}

func createTestEventWorkflowExecutionStarted(eventID int64, attr *s.WorkflowExecutionStartedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{EventId: common.Int64Ptr(eventID), EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionStarted), WorkflowExecutionStartedEventAttributes: attr}
}

func createTestEventActivityTaskScheduled(eventID int64, attr *s.ActivityTaskScheduledEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeActivityTaskScheduled),
		ActivityTaskScheduledEventAttributes: attr}
}

func createTestEventActivityTaskStarted(eventID int64, attr *s.ActivityTaskStartedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                            common.Int64Ptr(eventID),
		EventType:                          common.EventTypePtr(s.EventTypeActivityTaskStarted),
		ActivityTaskStartedEventAttributes: attr}
}

func createTestEventActivityTaskCompleted(eventID int64, attr *s.ActivityTaskCompletedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeActivityTaskCompleted),
		ActivityTaskCompletedEventAttributes: attr}
}

func createTestEventActivityTaskTimedOut(eventID int64, attr *s.ActivityTaskTimedOutEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                             common.Int64Ptr(eventID),
		EventType:                           common.EventTypePtr(s.EventTypeActivityTaskTimedOut),
		ActivityTaskTimedOutEventAttributes: attr}
}

func createTestEventDecisionTaskScheduled(eventID int64, attr *s.DecisionTaskScheduledEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeDecisionTaskScheduled),
		DecisionTaskScheduledEventAttributes: attr}
}

func createTestEventDecisionTaskStarted(eventID int64) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeDecisionTaskStarted)}
}

func createTestEventWorkflowExecutionSignaled(eventID int64, signalName string) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionSignaled),
		WorkflowExecutionSignaledEventAttributes: &s.WorkflowExecutionSignaledEventAttributes{
			SignalName: common.StringPtr(signalName),
			Identity:   common.StringPtr("test-identity"),
		},
	}
}

func createTestEventDecisionTaskCompleted(eventID int64, attr *s.DecisionTaskCompletedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeDecisionTaskCompleted),
		DecisionTaskCompletedEventAttributes: attr}
}

func createTestEventSignalExternalWorkflowExecutionFailed(eventID int64, attr *s.SignalExternalWorkflowExecutionFailedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeSignalExternalWorkflowExecutionFailed),
		SignalExternalWorkflowExecutionFailedEventAttributes: attr,
	}
}

func createWorkflowTask(
	events []*s.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
) *s.PollForDecisionTaskResponse {
	eventsCopy := make([]*s.HistoryEvent, len(events))
	copy(eventsCopy, events)
	return &s.PollForDecisionTaskResponse{
		PreviousStartedEventId: common.Int64Ptr(previousStartEventID),
		WorkflowType:           workflowTypePtr(WorkflowType{workflowName}),
		History:                &s.History{Events: eventsCopy},
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("fake-workflow-id"),
			RunId:      common.StringPtr(uuid.New()),
		},
	}
}

func createQueryTask(
	events []*s.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
	queryType string,
) *s.PollForDecisionTaskResponse {
	task := createWorkflowTask(events, previousStartEventID, workflowName)
	task.Query = &s.WorkflowQuery{
		QueryType: common.StringPtr(queryType),
	}
	return task
}

var testWorkflowTaskTasklist = "tl1"

func (t *TaskHandlersTestSuite) testWorkflowTaskWorkflowExecutionStartedHelper(params workerExecutionParameters) {
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &testWorkflowTaskTasklist}}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted() {
	params := workerExecutionParameters{
		TaskList: testWorkflowTaskTasklist,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted_WithDataConverter() {
	params := workerExecutionParameters{
		TaskList:      testWorkflowTaskTasklist,
		Identity:      "test-id-1",
		Logger:        t.logger,
		DataConverter: newTestDataConverter(),
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_ActivityTaskScheduled() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     &s.TaskList{Name: &taskList},
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
		createTestEventDecisionTaskStarted(8),
	}
	task := createWorkflowTask(testEvents[0:3], 0, "HelloWorld_Workflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response := request.(*s.RespondDecisionTaskCompletedRequest)

	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)

	// Schedule an activity and see if we complete workflow, Having only one last decision.
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, _, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response = request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeCompleteWorkflowExecution, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].CompleteWorkflowExecutionDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow_Sticky() {
	// Schedule an activity and see if we complete workflow.
	taskList := "sticky-tl"
	execution := &s.WorkflowExecution{
		WorkflowId: common.StringPtr("fake-workflow-id"),
		RunId:      common.StringPtr(uuid.New()),
	}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     &s.TaskList{Name: &taskList},
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())

	// first make progress on the workflow
	task := createWorkflowTask(testEvents[0:1], 0, "HelloWorld_Workflow")
	task.StartedEventId = common.Int64Ptr(1)
	task.WorkflowExecution = execution
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)

	// then check the current state using query task
	task = createQueryTask([]*s.HistoryEvent{}, 6, "HelloWorld_Workflow", "test-query")
	task.WorkflowExecution = execution
	queryResp, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NoError(err)
	t.verifyQueryResult(queryResp, "waiting-activity-result")
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow_NonSticky() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     &s.TaskList{Name: &taskList},
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
		createTestEventDecisionTaskStarted(8),
		createTestEventWorkflowExecutionSignaled(9, "test-signal"),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}

	// query after first decision task (notice the previousStartEventID is always the last eventID for query task)
	task := createQueryTask(testEvents[0:3], 3, "HelloWorld_Workflow", "test-query")
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	response, _, _ := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after activity task complete but before second decision task started
	task = createQueryTask(testEvents[0:7], 7, "HelloWorld_Workflow", "test-query")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	response, _, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after second decision task
	task = createQueryTask(testEvents[0:8], 8, "HelloWorld_Workflow", "test-query")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	response, _, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.verifyQueryResult(response, "done")

	// query after second decision task with extra events
	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", "test-query")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	response, _, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.verifyQueryResult(response, "done")

	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", "invalid-query-type")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	response, _, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NotNil(response)
	queryResp, ok := response.(*s.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.NotNil(queryResp.ErrorMessage)
	t.Contains(*queryResp.ErrorMessage, "unknown queryType")
}

func (t *TaskHandlersTestSuite) verifyQueryResult(response interface{}, expectedResult string) {
	t.NotNil(response)
	queryResp, ok := response.(*s.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.Nil(queryResp.ErrorMessage)
	t.NotNil(queryResp.QueryResult)
	encodedValue := newEncodedValue(queryResp.QueryResult, nil)
	var queryResult string
	err := encodedValue.Get(&queryResult)
	t.NoError(err)
	t.Equal(expectedResult, queryResult)
}

func (t *TaskHandlersTestSuite) TestCacheEvictionWhenErrorOccurs() {
	taskList := "taskList"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     &s.TaskList{Name: &taskList},
		}),
	}
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	// now change the history event so it does not match to decision produced via replay
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("some-other-activity")
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params)
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})

	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// There should be nothing in the cache.
	t.EqualValues(getWorkflowCache().Size(), 0)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_NondeterministicDetection() {
	taskList := "taskList"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     &s.TaskList{Name: &taskList},
		}),
	}
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	// there should be no error as the history events matched the decisions.
	t.NoError(err)
	t.NotNil(response)

	// now change the history event so it does not match to decision produced via replay
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("some-other-activity")
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params)
	request, _, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// now, create a new task handler with fail nondeterministic workflow policy
	// and verify that it handles the mismatching history correctly.
	params.NonDeterministicWorkflowPolicy = NonDeterministicWorkflowPolicyFailWorkflow
	failOnNondeterminismTaskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, _, err = failOnNondeterminismTaskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	// When FailWorkflow policy is set, task handler does not return an error,
	// because it will indicate non determinism in the request.
	t.NoError(err)
	// Verify that request is a RespondDecisionTaskCompleteRequest
	response, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	// Verify there's at least 1 decision
	// and the last last decision is to fail workflow
	// and contains proper justification.(i.e. nondeterminism).
	t.True(len(response.Decisions) > 0)
	closeDecision := response.Decisions[len(response.Decisions)-1]
	t.Equal(*closeDecision.DecisionType, s.DecisionTypeFailWorkflowExecution)
	t.Contains(*closeDecision.FailWorkflowExecutionDecisionAttributes.Reason, "NonDeterministicWorkflowPolicyFailWorkflow")

	// now with different package name to activity type
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("new-package.Greeter_Activity")
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, _, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NoError(err)
	t.NotNil(request)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowReturnsPanicError() {
	taskList := "taskList"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, "ReturnPanicWorkflow")
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.EqualValues(s.DecisionTypeFailWorkflowExecution, r.Decisions[0].GetDecisionType())
	attr := r.Decisions[0].FailWorkflowExecutionDecisionAttributes
	t.EqualValues("cadenceInternal:Panic", attr.GetReason())
	details := string(attr.Details)
	t.True(strings.HasPrefix(details, "\"panicError"), details)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowPanics() {
	taskList := "taskList"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, "PanicWorkflow")
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskFailedRequest)
	t.True(ok)
	t.EqualValues("WORKFLOW_WORKER_UNHANDLED_FAILURE", r.Cause.String())
	t.EqualValues("panicError", string(r.Details))
}

func (t *TaskHandlersTestSuite) TestGetWorkflowInfo() {
	taskList := "taskList"
	parentID := "parentID"
	parentRunID := "parentRun"
	cronSchedule := "5 4 * * *"
	continuedRunID := uuid.New()
	parentExecution := &s.WorkflowExecution{
		WorkflowId: &parentID,
		RunId:      &parentRunID,
	}
	parentDomain := "parentDomain"
	var attempt int32 = 123
	var executionTimeout int32 = 213456
	var taskTimeout int32 = 21
	workflowType := "GetWorkflowInfoWorkflow"
	lastCompletionResult, err := getDefaultDataConverter().ToData("lastCompletionData")
	t.NoError(err)
	startedEventAttributes := &s.WorkflowExecutionStartedEventAttributes{
		Input:                               lastCompletionResult,
		TaskList:                            &s.TaskList{Name: &taskList},
		ParentWorkflowExecution:             parentExecution,
		CronSchedule:                        &cronSchedule,
		ContinuedExecutionRunId:             &continuedRunID,
		ParentWorkflowDomain:                &parentDomain,
		Attempt:                             &attempt,
		ExecutionStartToCloseTimeoutSeconds: &executionTimeout,
		TaskStartToCloseTimeoutSeconds:      &taskTimeout,
		LastCompletionResult:                lastCompletionResult,
	}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, startedEventAttributes),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, workflowType)
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.EqualValues(s.DecisionTypeCompleteWorkflowExecution, r.Decisions[0].GetDecisionType())
	attr := r.Decisions[0].CompleteWorkflowExecutionDecisionAttributes
	var result WorkflowInfo
	t.NoError(getDefaultDataConverter().FromData(attr.Result, &result))
	t.EqualValues(taskList, result.TaskListName)
	t.EqualValues(parentID, result.ParentWorkflowExecution.ID)
	t.EqualValues(parentRunID, result.ParentWorkflowExecution.RunID)
	t.EqualValues(cronSchedule, *result.CronSchedule)
	t.EqualValues(continuedRunID, *result.ContinuedExecutionRunID)
	t.EqualValues(parentDomain, *result.ParentWorkflowDomain)
	t.EqualValues(attempt, result.Attempt)
	t.EqualValues(executionTimeout, result.ExecutionStartToCloseTimeoutSeconds)
	t.EqualValues(taskTimeout, result.TaskStartToCloseTimeoutSeconds)
	t.EqualValues(workflowType, result.WorkflowType.Name)
	t.EqualValues(testDomain, result.Domain)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_CancelActivityBeforeSent() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_WorkflowCancel")

	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task})
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	//t.printAllDecisions(response.Decisions)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeCompleteWorkflowExecution, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].CompleteWorkflowExecutionDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_PageToken() {
	// Schedule a decision activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	task.NextPageToken = []byte("token")

	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}

	nextEvents := []*s.HistoryEvent{
		createTestEventDecisionTaskStarted(3),
	}

	historyIterator := &historyIteratorImpl{
		iteratorFunc: func(nextToken []byte) (*s.History, []byte, error) {
			return &s.History{nextEvents}, nil, nil
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, getHostEnvironment())
	request, _, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task, historyIterator: historyIterator})
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NoError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	cancelRequested := false
	heartbeatResponse := s.RecordActivityTaskHeartbeatResponse{CancelRequested: &cancelRequested}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(&heartbeatResponse, nil)

	cadenceInvoker := &cadenceInvoker{
		identity:  "Test_Cadence_Invoker",
		service:   mockService,
		taskToken: nil,
	}

	heartbeatErr := cadenceInvoker.Heartbeat(nil)

	t.Nil(heartbeatErr)
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	entityNotExistsError := &s.EntityNotExistsError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(nil, entityNotExistsError)

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		func() {},
		0)

	heartbeatErr := cadenceInvoker.Heartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*s.EntityNotExistsError)
	t.True(ok, "heartbeatErr must be EntityNotExistsError.")
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithDomainNotActiveError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	domainNotActiveError := &s.DomainNotActiveError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(nil, domainNotActiveError)

	called := false
	cancelHandler := func() { called = true }

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		cancelHandler,
		0)

	heartbeatErr := cadenceInvoker.Heartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*s.DomainNotActiveError)
	t.True(ok, "heartbeatErr must be DomainNotActiveError.")
	t.True(called)
}

type testActivityDeadline struct {
	logger *zap.Logger
	d      time.Duration
}

func (t *testActivityDeadline) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if d, _ := ctx.Deadline(); d.IsZero() {
		panic("invalid deadline provided")
	}
	if t.d != 0 {
		// Wait till deadline expires.
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return nil, nil
}

func (t *testActivityDeadline) ActivityType() ActivityType {
	return ActivityType{Name: "test"}
}

func (t *testActivityDeadline) GetFunction() interface{} {
	return t.Execute
}

type deadlineTest struct {
	actWaitDuration  time.Duration
	ScheduleTS       time.Time
	ScheduleDuration int32
	StartTS          time.Time
	StartDuration    int32
	err              error
}

func (t *TaskHandlersTestSuite) TestActivityExecutionDeadline() {
	deadlineTests := []deadlineTest{
		{time.Duration(0), time.Now(), 3, time.Now(), 3, nil},
		{time.Duration(0), time.Now(), 4, time.Now(), 3, nil},
		{time.Duration(0), time.Now(), 3, time.Now(), 4, nil},
		{time.Duration(0), time.Now().Add(-1 * time.Second), 1, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(0), time.Now(), 1, time.Now().Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{time.Duration(0), time.Now().Add(-1 * time.Second), 1, time.Now().Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 1, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 2, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 1, time.Now(), 2, context.DeadlineExceeded},
	}
	a := &testActivityDeadline{logger: t.logger}
	hostEnv := getHostEnvironment()
	hostEnv.addActivity(a.ActivityType().Name, a)

	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	for i, d := range deadlineTests {
		a.d = d.actWaitDuration
		wep := workerExecutionParameters{
			Logger:        t.logger,
			DataConverter: getDefaultDataConverter(),
		}
		activityHandler := newActivityTaskHandler(mockService, wep, hostEnv)
		pats := &s.PollForActivityTaskResponse{
			TaskToken: []byte("token"),
			WorkflowExecution: &s.WorkflowExecution{
				WorkflowId: common.StringPtr("wID"),
				RunId:      common.StringPtr("rID")},
			ActivityType:                  &s.ActivityType{Name: common.StringPtr("test")},
			ActivityId:                    common.StringPtr(uuid.New()),
			ScheduledTimestamp:            common.Int64Ptr(d.ScheduleTS.UnixNano()),
			ScheduleToCloseTimeoutSeconds: common.Int32Ptr(d.ScheduleDuration),
			StartedTimestamp:              common.Int64Ptr(d.StartTS.UnixNano()),
			StartToCloseTimeoutSeconds:    common.Int32Ptr(d.StartDuration),
			WorkflowType: &s.WorkflowType{
				Name: common.StringPtr("wType"),
			},
			WorkflowDomain: common.StringPtr("domain"),
		}
		td := fmt.Sprintf("testIndex: %v, testDetails: %v", i, d)
		r, err := activityHandler.Execute(tasklist, pats)
		t.logger.Info(fmt.Sprintf("test: %v, result: %v err: %v", td, r, err))
		t.Equal(d.err, err, td)
		if err != nil {
			t.Nil(r, td)
		}
	}
}

func Test_NonDeterministicCheck(t *testing.T) {
	decisionTypes := s.DecisionType_Values()
	require.Equal(t, 12, len(decisionTypes), "If you see this error, you are adding new decision type. "+
		"Before updating the number to make this test pass, please make sure you update isDecisionMatchEvent() method "+
		"to check the new decision type. Otherwise the replay will fail on the new decision event.")

	eventTypes := s.EventType_Values()
	decisionEventTypeCount := 0
	for _, et := range eventTypes {
		if isDecisionEvent(et) {
			decisionEventTypeCount++
		}
	}
	// CancelTimer has 2 corresponding events.
	require.Equal(t, len(decisionTypes)+1, decisionEventTypeCount, "Every decision type must have one matching event type. "+
		"If you add new decision type, you need to update isDecisionEvent() method to include that new event type as well.")
}
