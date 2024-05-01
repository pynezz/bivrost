// https://github.com/bradleyjkemp/sigma-go/blob/main/condition_parser.go

package sigma

import (
	"github.com/bradleyjkemp/sigma-go"
)

func test() {
	// You can load/create rules dynamically or use sigmac to load Sigma rule files
	var rule, _ = sigma.ParseRule(contents)

	// Rules need to be wrapped in an evaluator.
	// This is also where (if needed) you provide functions implementing the count, max, etc. aggregation functions
	e := sigma.Evaluator(rule, options...)

	// Get a stream of events from somewhere e.g. audit logs
	for event := range events {
		if e.Matches(ctx, event) {
			// Raise your alert here
		}
	}
}
