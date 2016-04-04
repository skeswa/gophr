package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSemverSelector(t *testing.T) {
	var (
		err    error
		semver SemverSelector
	)

	semver, err = NewSemverSelector("6", "1", "", "", "", "", "")
	assert.NotNil(t, err, "should fail on illegal prefixes")

	semver, err = NewSemverSelector("", "1", "", "", "", "", "?")
	assert.NotNil(t, err, "should fail on illegal suffixes")

	semver, err = NewSemverSelector("", "c", "", "", "", "", "")
	assert.NotNil(t, err, "should fail on illegal major segment")

	semver, err = NewSemverSelector("", "", "", "", "", "", "")
	assert.NotNil(t, err, "should fail on no major segment provided")

	semver, err = NewSemverSelector("", "1", "", "1", "", "", "")
	assert.NotNil(t, err, "should fail on gap between version segments")

	semver, err = NewSemverSelector("", "1", "z", "", "", "", "")
	assert.NotNil(t, err, "should fail on illegal minor segment")

	semver, err = NewSemverSelector("", "1", "x", "1", "", "", "")
	assert.NotNil(t, err, "should fail on an segment trailing a wildcard")

	semver, err = NewSemverSelector("", "1", "x", "x", "", "", "")
	assert.NotNil(t, err, "should fail on an segment trailing a wildcard")

	semver, err = NewSemverSelector("~", "1", "x", "", "", "", "")
	assert.NotNil(t, err, "should fail when prefix is mixed with minor wildcard")

	semver, err = NewSemverSelector("~", "1", "1", "x", "", "", "")
	assert.NotNil(t, err, "should fail when prefix is mixed with patch wildcard")

	semver, err = NewSemverSelector("", "1", "1", "x", "alpha", "", "")
	assert.NotNil(t, err, "should fail on an segment trailing a wildcard")

	semver, err = NewSemverSelector("~", "1", "1", "z", "", "", "")
	assert.NotNil(t, err, "should fail on illegal patch segment")

	semver, err = NewSemverSelector("~", "1", "1", "", "alpha", "", "")
	assert.NotNil(t, err, "should fail on gap between version segments")

	semver, err = NewSemverSelector("~", "1", "1", "1", "alpha", "x", "")
	assert.NotNil(t, err, "should fail when prefix is mixed with prerelease wildcard")

	semver, err = NewSemverSelector("~", "1", "1", "", "", "x", "")
	assert.NotNil(t, err, "should fail on gap between version segments")

	semver, err = NewSemverSelector("~", "1", "1", "1", "alpha", "z", "")
	assert.NotNil(t, err, "should fail on illegal prelease segment")

	semver, err = NewSemverSelector("~", "1", "2", "", "", "", "+")
	assert.NotNil(t, err, "should fail when prefix is mixed with suffix")

	semver, err = NewSemverSelector("", "1", "2", "x", "", "", "+")
	assert.NotNil(t, err, "should fail when wildcard is mixed with suffix")

	semver, err = NewSemverSelector("", "1", "2", "x", "", "", "x")
	assert.NotNil(t, err, "should fail when wildcard is mixed with suffix")

	// semver, err = NewSemverSelector("~", "1", "", "", "", "", "")
	// assert.NotNil(t, err)

	semver, err = NewSemverSelector("", "1", "", "", "", "", "")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixNone, semver.Prefix, "prefix should be unspecified")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 1, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PatchVersion.Type, "patch should be type unspecified")
	assert.Equal(t, "", semver.PrereleaseLabel, "prerelease label should be empty")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PrereleaseVersion.Type, "prerelease should be type unspecified")
	assert.Equal(t, semverSelectorSuffixNone, semver.Suffix, "suffix should be unspecified")

	semver, err = NewSemverSelector("", "2", "", "", "", "", "-")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixNone, semver.Prefix, "prefix should be unspecified")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 2, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PatchVersion.Type, "patch should be type unspecified")
	assert.Equal(t, "", semver.PrereleaseLabel, "prerelease label should be empty")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PrereleaseVersion.Type, "prerelease should be type unspecified")
	assert.Equal(t, semverSelectorSuffixLessThan, semver.Suffix, "suffix should be less than")

	semver, err = NewSemverSelector("~", "1", "2", "", "", "", "")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixTilde, semver.Prefix, "prefix should be a tilde")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 1, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, 2, semver.MinorVersion.Number, "minor should be the correct number")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PatchVersion.Type, "patch should be type unspecified")
	assert.Equal(t, "", semver.PrereleaseLabel, "prerelease label should be empty")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PrereleaseVersion.Type, "prerelease should be type unspecified")
	assert.Equal(t, semverSelectorSuffixNone, semver.Suffix, "suffix should be unspecified")

	semver, err = NewSemverSelector("^", "1", "2", "3", "", "", "")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixCarat, semver.Prefix, "prefix should be a carat")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 1, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, 2, semver.MinorVersion.Number, "minor should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.PatchVersion.Type, "patch should be type number")
	assert.Equal(t, 3, semver.PatchVersion.Number, "patch should be the correct number")
	assert.Equal(t, "", semver.PrereleaseLabel, "prerelease label should be empty")
	assert.Equal(t, semverSegmentTypeUnspecified, semver.PrereleaseVersion.Type, "prerelease should be type unspecified")
	assert.Equal(t, semverSelectorSuffixNone, semver.Suffix, "suffix should be unspecified")

	semver, err = NewSemverSelector("", "1", "2", "3", "alpha", "x", "")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixNone, semver.Prefix, "prefix should be unspecified")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 1, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, 2, semver.MinorVersion.Number, "minor should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.PatchVersion.Type, "patch should be type number")
	assert.Equal(t, 3, semver.PatchVersion.Number, "patch should be the correct number")
	assert.Equal(t, "alpha", semver.PrereleaseLabel, "prerelease label should be alpha")
	assert.Equal(t, semverSegmentTypeWildcard, semver.PrereleaseVersion.Type, "prerelease should be type wildcard")
	assert.Equal(t, semverSelectorSuffixNone, semver.Suffix, "suffix should be unspecified")

	semver, err = NewSemverSelector("", "1", "2", "3", "beta", "43", "+")
	assert.Nil(t, err)
	assert.Equal(t, semverSelectorPrefixNone, semver.Prefix, "prefix should be unspecified")
	assert.Equal(t, semverSegmentTypeNumber, semver.MajorVersion.Type, "major should be type number")
	assert.Equal(t, 1, semver.MajorVersion.Number, "major should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.MinorVersion.Type, "minor should be type number")
	assert.Equal(t, 2, semver.MinorVersion.Number, "minor should be the correct number")
	assert.Equal(t, semverSegmentTypeNumber, semver.PatchVersion.Type, "patch should be type number")
	assert.Equal(t, 3, semver.PatchVersion.Number, "patch should be the correct number")
	assert.Equal(t, "beta", semver.PrereleaseLabel, "prerelease label should be alpha")
	assert.Equal(t, semverSegmentTypeNumber, semver.PrereleaseVersion.Type, "prerelease should be type number")
	assert.Equal(t, 43, semver.PrereleaseVersion.Number, "prerelease should the correct number")
	assert.Equal(t, semverSelectorSuffixGreaterThan, semver.Suffix, "suffix should be greater than")
}

func TestSemverString(t *testing.T) {
	var (
		semver SemverSelector
	)

	semver, _ = NewSemverSelector("", "1", "", "", "", "", "")
	assert.Equal(t, "1", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("~", "1", "", "", "", "", "")
	assert.Equal(t, "~1", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("^", "1", "", "", "", "", "")
	assert.Equal(t, "^1", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("~", "1", "2", "", "", "", "")
	assert.Equal(t, "~1.2", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "", "", "")
	assert.Equal(t, "1.2.3", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "x", "", "", "", "")
	assert.Equal(t, "1.x", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "x", "", "", "")
	assert.Equal(t, "1.2.x", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "alpha", "", "")
	assert.Equal(t, "1.2.3-alpha", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "alpha", "x", "")
	assert.Equal(t, "1.2.3-alpha.x", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "alpha", "4", "")
	assert.Equal(t, "1.2.3-alpha.4", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "alpha", "4", "+")
	assert.Equal(t, "1.2.3-alpha.4+", semver.String(), "serialized semver should match expectations")

	semver, _ = NewSemverSelector("", "1", "2", "3", "alpha", "4", "-")
	assert.Equal(t, "1.2.3-alpha.4-", semver.String(), "serialized semver should match expectations")

	semver = SemverSelector{
		MajorVersion: SemverSelectorSegment{
			Type: semverSegmentTypeWildcard,
		},
	}
	assert.Panics(t, func() {
		_ = semver.String()
	}, "semver serialization should fail since major segment in an incorrect type")

	semver = SemverSelector{
		MajorVersion: SemverSelectorSegment{
			Type: semverSegmentTypeUnspecified,
		},
	}
	assert.Panics(t, func() {
		_ = semver.String()
	}, "semver serialization should fail since major segment in an incorrect type")
}