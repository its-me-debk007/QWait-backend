package util

const NEAR_BY_DISTANCE = 2.5

type Category int

const (
	HOSPITAL Category = iota
	BANK
	REGCAMP
	GOVT_OFFICE
	TICKETING_SYSTEM
)

func (category Category) String() string {
	switch category {
	case HOSPITAL:
		return "hospital"
	case BANK:
		return "bank"
	case REGCAMP:
		return "registration_camp"
	case GOVT_OFFICE:
		return "govt_office"
	case TICKETING_SYSTEM:
		return "ticketing_system"
	}

	return ""
}
