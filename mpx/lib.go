package mpx

func StringParamDefault(param string, def string) string {
	if param == "" {
		return def
	}
	return param
}
