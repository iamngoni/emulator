package main

type EmulatorConfig struct {
	Name         string
	DeviceType   string
	APILevel     string
	SystemImage  string
	RAM          int
	Resolution   string
	Density      int
	HasKeyboard  bool
	ExtraConfigs []string
}
