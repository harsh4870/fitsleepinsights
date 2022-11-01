// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import "time"

// /activities/%s/date/%s/%s.json

type ActivityCaloriesSeries struct {
	TimeSeries []TimeStep `json:"activities-activityCalories"`
}

type TimeStep struct {
	DateTime time.Time `json:"dateTime"`
	Value    string    `json:"value"`
}

// /activities/%s/date/%s/%s.json

type CaloriesSeries struct {
	TimeSeries []TimeStep `json:"activities-calories"`
}

// /activities/%s/date/%s/%s.json

type CaloriesBMRSeries struct {
	TimeSeries []TimeStep `json:"activities-caloriesBMR"`
}

// /activities/%s/date/%s/%s.json

type DistanceSeries struct {
	TimeSeries []TimeStep `json:"activities-distance"`
}

// /activities/%s/date/%s/%s.json

type ElevationSeries struct {
	TimeSeries []TimeStep `json:"activities-elevation"`
}

// /activities/%s/date/%s/%s.json

type FloorsSeries struct {
	TimeSeries []TimeStep `json:"activities-floors"`
}

// /activities/%s/date/%s/%s.json

type MinutesSedentarySeries struct {
	TimeSeries []TimeStep `json:"activities-minutesSedentary"`
}

// /activities/%s/date/%s/%s.json

type MinutesLightlyActiveSeries struct {
	TimeSeries []TimeStep `json:"activities-minutesLightlyActive"`
}

// /activities/%s/date/%s/%s.json

type MinutesFailryActiveSeries struct {
	TimeSeries []TimeStep `json:"activities-minutesFairlyActive"`
}

// /activities/%s/date/%s/%s.json

type MinutesVeryActiveSeries struct {
	TimeSeries []TimeStep `json:"activities-minutesVeryActive"`
}

// /activities/%s/date/%s/%s.json

type StepsSeries struct {
	TimeSeries []TimeStep `json:"activities-steps"`
}