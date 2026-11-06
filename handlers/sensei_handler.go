package handlers

import "net/http"

type (
	ISenseiHandler interface {
		AssignSenseiClass(w http.ResponseWriter, r *http.Request)
		GetSenseiClassSchedule(w http.ResponseWriter, r *http.Request)
		GetSenseiClassScheduleDetail(w http.ResponseWriter, r *http.Request)
		ListOfStudentRecords(w http.ResponseWriter, r *http.Request)
	}
)
