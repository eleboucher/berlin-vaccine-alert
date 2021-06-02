package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

type RHelios struct {
	Purposes []Purpose `json:"purposes"`
}

type Purpose struct {
	UUID                       string      `json:"uuid"`
	CreatedBy                  interface{} `json:"createdBy"`
	CreatedOn                  string      `json:"createdOn"`
	ChangedBy                  interface{} `json:"changedBy"`
	ChangedOn                  string      `json:"changedOn"`
	ObjectStatus               string      `json:"objectStatus"`
	OID                        int64       `json:"oid"`
	Name                       string      `json:"name"`
	ID                         string      `json:"id"`
	Eid                        interface{} `json:"eid"`
	PurposeGroupUUID           string      `json:"purposeGroupUUID"`
	PurposeGroupName           string      `json:"purposeGroupName"`
	PurposeDurationName        string      `json:"purposeDurationName"`
	DepartmentName             string      `json:"departmentName"`
	DepartmentUUID             string      `json:"departmentUUID"`
	SpecialtyUUID              string      `json:"specialtyUUID"`
	Flags                      int64       `json:"flags"`
	PurposeCategoryUUID        string      `json:"purposeCategoryUUID"`
	InternalInstructions       string      `json:"internalInstructions"`
	PhoneInstructions          interface{} `json:"phoneInstructions"`
	Description                interface{} `json:"description"`
	RequestInstructions        interface{} `json:"requestInstructions"`
	PreBookingInstructions     string      `json:"preBookingInstructions"`
	PostBookingInstructions    string      `json:"postBookingInstructions"`
	SMSInstructions            interface{} `json:"smsInstructions"`
	StyleUUID                  interface{} `json:"styleUUID"`
	MinParticipants            int64       `json:"minParticipants"`
	MaxParticipants            int64       `json:"maxParticipants"`
	StereotypeUUID             string      `json:"stereotypeUUID"`
	BookingPlanUUID            interface{} `json:"bookingPlanUUID"`
	BookingWindowUUID          string      `json:"bookingWindowUUID"`
	NotificationProfileUUID    interface{} `json:"notificationProfileUUID"`
	ProgressStepDescriptorUUID interface{} `json:"progressStepDescriptorUuid"`
	AttributeDescriptors       interface{} `json:"attributeDescriptors"`
	ProgressSteps              interface{} `json:"progressSteps"`
}

// Helios holds the information for fetching the information for the
// https://patienten.helios-gesundheit.de/ website
type Helios struct {
	resultSendLastAt time.Time
	lastResult       []*Result
}

const tHelios = "appointments for biontech available call https://patienten.helios-gesundheit.de/appointments/book-appointment?facility=10&physician=21646&purpose=33239&resource=58"

// Fetch fetches all the available appointment and filter then and return the results
func (h *Helios) Fetch() ([]*Result, error) {
	url := "https://api.patienten.helios-gesundheit.de/api/appointment/resources/21646/purposes?insuranceTypeId=1&specialtyUUID=c619bfb1-9e18-404d-b960-dfac6c072490"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	var resp RHelios
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Purposes) > 0 && resp.Purposes[0].BookingPlanUUID != nil {
		var ret Result
		ret.VaccineName = Biontech
		ret.Message = tHelios
		return []*Result{&ret}, nil
	}
	return nil, nil
}

// ShouldSendResult check if the result should be send now
func (h *Helios) ShouldSendResult(result []*Result) bool {
	if !reflect.DeepEqual(h.lastResult, result) && h.resultSendLastAt.Before(time.Now().Add(-1*time.Minute)) {
		return true
	}
	if h.resultSendLastAt.Before(time.Now().Add(-10 * time.Minute)) {
		return true
	}
	return false
}

// ResultSentNow set that the appointment has been sent
func (h *Helios) ResultSentNow(result []*Result) {
	h.resultSendLastAt = time.Now()
	h.lastResult = result
}
