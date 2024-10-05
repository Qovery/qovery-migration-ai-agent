package sources

type AppConfig interface {
	App() map[string]interface{}
	Name() string
	Cost() float64
	Map() map[string]interface{}
}
