package giterminism_inspector

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
)

const giterminismDocPageURL = "https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html"

var (
	LooseGiterminism         bool
	NonStrict                bool
	ReportedUncommittedPaths []string
)

type InspectionOptions struct {
	LooseGiterminism bool
	NonStrict        bool
}

func Init(opts InspectionOptions) error {
	LooseGiterminism = opts.LooseGiterminism
	NonStrict = opts.NonStrict
	return nil
}

func ReportUncommittedFile(ctx context.Context, path string) error {
	for _, p := range ReportedUncommittedPaths {
		if p == path {
			return nil
		}
	}
	ReportedUncommittedPaths = append(ReportedUncommittedPaths, path)

	if NonStrict {
		logboek.Context(ctx).Warn().LogF("WARNING: Uncommitted file %s was not taken into account (more info %s)\n", path, giterminismDocPageURL)
		return nil
	} else {
		return fmt.Errorf("restricted usage of uncommitted file %s (more info %s)", path, giterminismDocPageURL)
	}
}

func ReportMountDirectiveUsage(ctx context.Context) error {
	return fmt.Errorf("'mount' directive is forbidden due to enabled giterminism mode (more info %s), it is recommended to avoid this directive", giterminismDocPageURL)
}

func ReportGoTemplateEnvFunctionUsage(ctx context.Context, functionName string) error {
	return fmt.Errorf("go templates function %q is forbidden due to enabled giterminism mode (more info %s)", functionName, giterminismDocPageURL)
}

func PrintInspectionDebrief(ctx context.Context) {
	if NonStrict {
		if len(ReportedUncommittedPaths) > 0 {
			logboek.Context(ctx).Warn().LogLn()
			logboek.Context(ctx).Warn().LogF("### Giterminism inspection debrief ###\n")
			logboek.Context(ctx).Warn().LogLn()
			logboek.Context(ctx).Warn().LogF("Following uncommitted files were not taken into account:\n")
			for _, path := range ReportedUncommittedPaths {
				logboek.Context(ctx).Warn().LogF(" - %s\n", path)
			}
			logboek.Context(ctx).Warn().LogLn()
			logboek.Context(ctx).Warn().LogF("More info about giterminism in the werf avaiable on the page: %s\n", giterminismDocPageURL)
			logboek.Context(ctx).Warn().LogLn()
		}
	}
}