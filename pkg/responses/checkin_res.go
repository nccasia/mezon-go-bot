package responses

type CheckinRes struct {
	AccountEmployeeID       string  `json:"accountEmployeeId"`
	EmployeeID              string  `json:"employeeId"`
	FacialRecognitionStatus string  `json:"facialRecognitionStatus"`
	FirstName               string  `json:"firstName"`
	ImageVerifyID           string  `json:"imageVerifyId"`
	LastName                string  `json:"lastName"`
	Probability             float64 `json:"probability"`
	ShowMessage             bool    `json:"showMessage"`
	IdentityVerified        bool    `json:"identityVerified"`
}
