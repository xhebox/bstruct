package bstruct

import "fmt"

type FieldFlag byte

const (
	FlagSkipr FieldFlag = 0x1
	FlagSkipw FieldFlag = 0x2
	FlagSkip  FieldFlag = FlagSkipr | FlagSkipw
	// short for custom endian
	FlagCusEnd FieldFlag = 0x4
	FlagBig    FieldFlag = 0x8
)

func (this FieldFlag) String() string {
	skip := ""
	endian := "host"

	if this&FlagSkipr != 0 {
		skip += "r"
	}

	if this&FlagSkipw != 0 {
		skip += "w"
	}

	if this&FlagCusEnd != 0 {
		if this&FlagBig != 0 {
			endian = "big"
		} else {
			endian = "little"
		}
	}

	return fmt.Sprintf("{skip: %s, endian: %s}", skip, endian)
}
