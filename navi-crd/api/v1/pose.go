package v1

type Position struct {
	X string `json:"x,omitempty"`
	Y string `json:"y,omitempty"`
	Z string `json:"z,omitempty"`
}

type Orientation struct {
	W string `json:"w,omitempty"`
}

type InitialPose struct {
	Position    Position    `json:"position,omitempty"`
	Orientation Orientation `json:"orientation,omitempty"`
}

type GoalPose struct {
	Position    Position    `json:"position,omitempty"`
	Orientation Orientation `json:"orientation,omitempty"`
}
