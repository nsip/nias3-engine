// monitor-pipeline.go

package pipelines

import "log"

// WaitForPipeline waits for results from all error channels.
// It returns early on the first error.
func MonitorPipeline(pipelineName string, errs ...<-chan error) error {

	errc := MergeErrors(errs...)
	for err := range errc {
		if err != nil {
			log.Printf("error received in pipeline monitor (%s):\n\t%s\n", pipelineName, err)
			// return err
		}
	}
	log.Printf("monitored pipleine (%s) shutting down", pipelineName)
	return nil
}
