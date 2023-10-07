package workerhandler

import (
	"encoding/json"
	"fmt"
	"github.com/dev-sareno/ginamus/codec"
	"github.com/dev-sareno/ginamus/context"
	"github.com/dev-sareno/ginamus/dns"
	"github.com/dev-sareno/ginamus/dto"
	"log"
	"os"
	"strings"
)

func HandleJob(ctx context.WorkerContext, data []byte) {
	log.Printf("Received a message: %s", data)

	var job dto.Job
	if err := json.Unmarshal(data, &job); err != nil {
		log.Printf("invalid job input %s\n", err)
		return
	}
	if job.Data.Type != 0 {
		log.Printf("unsuported job type %d\n", job.Data.Type)
		return
	}

	// assign job to context
	ctx.Job = &job
	result, err := handleDnsResolution(ctx)
	if err != nil {
		log.Printf("dns resolution failed. %s\n", err)
		return
	}

	log.Printf("dns resolution result: %s\n", result)
}

func handleDnsResolution(ctx context.WorkerContext) (string, error) {
	lookupType := os.Getenv("WORKER_DNS_LOOKUP_TYPE")
	switch lookupType {
	case "A":
		const activityId = "lookup-a"

		job := ctx.Job
		job.LastActivityId = activityId // set activity id

		jobInput := job.Data.Input

		jobOutput := dto.ActivityOutput{
			Index:   int32(len(job.Data.Outputs)),
			Id:      activityId,
			Data:    "",
			IsOk:    true,
			Message: job.LastActivityMessage,
		}
		var lookupResult []string // list of the resolved values
		hasWarning := false
		hasError := false
		for _, v := range jobInput.Domains {
			lookup := dns.IpResolver{}
			lookup.SetValue(v)
			result, err := lookup.Resolve()
			if err != nil {
				// lookup failed
				log.Printf("a lookup failed. %s\n", err)
				lookupResult = append(lookupResult, "") // append empty
				hasError = true
			} else {
				// lookup successful
				// ipv4 lookup is expected length of one since it  doesn't have a child
				if len(result) != 1 {
					log.Println("WARNING: ipresolver is expected to return a length of 1")
					// this is invalid, consider as failed
					lookupResult = append(lookupResult, "") // append empty
					hasWarning = true
					continue
				}
				lookupResult = append(lookupResult, result[0].Value)
			}
		}

		var msg string
		if hasWarning {
			msg = "completed with warning"
		} else if hasError {
			msg = "completed with errors"
		} else {
			msg = "completed"
		}

		// encode result
		b, _ := json.Marshal(&lookupResult)

		// finalize job output
		jobOutput.Data = string(b)
		jobOutput.Message = msg
		jobOutput.IsOk = true
		job.LastActivityMessage = msg
		job.LastActivityIsOk = true

		job.Data.Outputs = append(job.Data.Outputs, jobOutput)

		codec.Encode(job)
		break
	case "CNAME":
		log.Println("TODO: implement cname lookup")
	default:
		return "", fmt.Errorf("invalid dns lookup type %s\n", lookupType)
	}
	return "", nil
}

func test() {
	domain := "github.com"

	ipResolver := dns.IpResolver{}
	cnameResolver := dns.CnameResolver{Child: &ipResolver}
	recordResolver := dns.RecordResolver{Child: &cnameResolver}
	recordResolver.SetValue(domain)
	result, err := recordResolver.Resolve()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var items []string
	for _, item := range result {
		items = append(items, item.Value)
	}
	fmt.Printf("%v\n", strings.Join(items, "\n"))
}
