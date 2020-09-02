package main

type config struct {
	Features features `yaml:"features"`
}

type features struct {
	Enabled      bool `yaml:"enabled"`
	IdleDuration int  `yaml:"idle_duration"`
}
