// worker/cmd/main.go
package main

import (
	"context"
	"log"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var (
	TemporalAddress        string = "temporal-dev:7233"
	TaskQueue              string = "default-activity-queue"
	ChildWorkflowTaskQueue string = "child-activity-queue"
	WorkflowName           string = "FourActivityWorkflow"
	ChildWorkflowName      string = "Activity2ChildWorkflow"
)

// Workflow definition
func FourActivityWorkflow(ctx workflow.Context, name string) (string, error) {

	aoDefault := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
	}
	ctxDefault := workflow.WithActivityOptions(ctx, aoDefault)

	results := []string{}


	var r1 string
	if err := workflow.ExecuteActivity(ctxDefault, Activity1, name+"-1").Get(ctxDefault, &r1); err != nil {
		return "", err
	}
	results = append(results, r1)

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue: ChildWorkflowTaskQueue,
	}
	ctxChild := workflow.WithChildOptions(ctx, cwo)
	var r2 string
	if err := workflow.ExecuteChildWorkflow(ctxChild, ChildWorkflowName, name+"-2").Get(ctxChild, &r2); err != nil {
		return "", err
	}
	results = append( results, r2)

	// Activity3 (default queue)
	var r3 string
	if err := workflow.ExecuteActivity(ctxDefault, Activity3, name+"-3").Get(ctxDefault, &r3); err != nil {
		return "", err
	}
	results = append(results, r3)

	// Activity4 (default queue)
	var r4 string
	if err := workflow.ExecuteActivity(ctxDefault, Activity4, name+"-4").Get(ctxDefault, &r4); err != nil {
		return "", err
	}
	results = append(results, r4)

	return strings.Join(results, ","), nil
}

func ChildWorkflowForActivity2(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		TaskQueue:           ChildWorkflowTaskQueue, // <--- add this
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	if err := workflow.ExecuteActivity(ctx, Activity2, name).Get(ctx, &result); err != nil {
		return "", err
	}
	return result, nil
}

// Activities handled by this worker (Activity1, Activity3, Activity4)
func Activity1(ctx context.Context, name string) (string, error) {
	time.Sleep(1 * time.Second)
	return "Activity1-" + name, nil
}

func Activity2(ctx context.Context, name string) (string, error) {
	time.Sleep(1 * time.Second)
	return "Activity2-" + name, nil
}

func Activity3(ctx context.Context, name string) (string, error) {
	time.Sleep(1 * time.Second)
	return "Activity3-" + name, nil
}

func Activity4(ctx context.Context, name string) (string, error) {
	time.Sleep(1 * time.Second)
	return "Activity4-" + name, nil
}

func main() {

	c, err := client.Dial(client.Options{HostPort: TemporalAddress})
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	// Worker for workflow + Activities 1, 3, 4
	w := worker.New(c, TaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(FourActivityWorkflow, workflow.RegisterOptions{
		Name: WorkflowName,
	})
	w.RegisterActivity(Activity1)
	w.RegisterActivity(Activity3)
	w.RegisterActivity(Activity4)

	childWorker := worker.New(c, ChildWorkflowTaskQueue, worker.Options{})
	childWorker.RegisterWorkflowWithOptions(ChildWorkflowForActivity2, workflow.RegisterOptions{
		Name: ChildWorkflowName,
	})
	childWorker.RegisterActivity(Activity2)
	go childWorker.Run(worker.InterruptCh())

	log.Println("Parent worker started (Workflow + Activities 1,3,4)...")
	log.Println("Child worker started (Child Workflow + Activity2)...")

	w.Run(worker.InterruptCh())

}
