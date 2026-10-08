package models

// TimeSlot represents a time slot for doctor availability
// @Description Time slot information showing doctor's available appointment times
type TimeSlot struct {
	Date      string `json:"date" example:"2025-11-15"`
	StartTime string `json:"startTime" example:"10:00:00"`
	EndTime   string `json:"endTime" example:"10:30:00"`
	Available bool   `json:"available" example:"true"`
}
