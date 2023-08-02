// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fileconsumer // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"

import (
	"regexp"

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/multierr"
)

type MatchingCriteria struct {
	Include          []string         `mapstructure:"include,omitempty"`
	Exclude          []string         `mapstructure:"exclude,omitempty"`
	OrderingCriteria OrderingCriteria `mapstructure:"ordering_criteria,omitempty"`
}

type OrderingCriteria struct {
	Regex  string         `mapstructure:"regex,omitempty"`
	SortBy []sortRuleImpl `mapstructure:"sort_by,omitempty"`
}

type NumericSortRule struct {
	baseSortRule `mapstructure:",squash"`
}

type AlphabeticalSortRule struct {
	baseSortRule `mapstructure:",squash"`
}

type TimestampSortRule struct {
	baseSortRule `mapstructure:",squash"`
	Layout       string `mapstructure:"layout,omitempty"`
	Location     string `mapstructure:"location,omitempty"`
}

type baseSortRule struct {
	RegexKey  string `mapstructure:"regex_key,omitempty"`
	Ascending bool   `mapstructure:"ascending,omitempty"`
	SortType  string `mapstructure:"sort_type,omitempty"`
}

type sortRuleImpl struct {
	sortRule
}

// findFiles gets a list of paths given an array of glob patterns to include and exclude
func (f MatchingCriteria) findFiles() ([]string, error) {
	all := make([]string, 0, len(f.Include))
	for _, include := range f.Include {
		matches, _ := doublestar.FilepathGlob(include, doublestar.WithFilesOnly()) // compile error checked in build
	INCLUDE:
		for _, match := range matches {
			for _, exclude := range f.Exclude {
				if itMatches, _ := doublestar.PathMatch(exclude, match); itMatches {
					continue INCLUDE
				}
			}

			for _, existing := range all {
				if existing == match {
					continue INCLUDE
				}
			}

			all = append(all, match)
		}
	}

	if len(all) == 0 || len(f.OrderingCriteria.SortBy) == 0 {
		return all, nil
	}

	re := regexp.MustCompile(f.OrderingCriteria.Regex)

	var errs error
	for _, SortPattern := range f.OrderingCriteria.SortBy {
		sortedFiles, err := SortPattern.sort(re, all)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		all = sortedFiles
	}

	return []string{all[0]}, errs
}