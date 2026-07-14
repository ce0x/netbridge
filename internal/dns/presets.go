package dns

var DefaultPresets = []struct {
	Name    string
	Servers []string
}{
	{Name: "cloudflare", Servers: []string{"1.1.1.1", "1.0.0.1"}},
	{Name: "google", Servers: []string{"8.8.8.8", "8.8.4.4"}},
	{Name: "quad9", Servers: []string{"9.9.9.9", "149.112.112.112"}},
	{Name: "adguard", Servers: []string{"94.140.14.14", "94.140.15.15"}},
	{Name: "system", Servers: []string{}},
}
