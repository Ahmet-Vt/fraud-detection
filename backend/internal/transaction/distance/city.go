package distance

import "strings"

type CityCoordinates struct {
	Lat float64
	Lon float64
}

var CityCoordinatesByName = map[string]CityCoordinates{
	"adana":          {Lat: 37, Lon: 35.3213333},
	"adıyaman":       {Lat: 37.7641667, Lon: 38.2761667},
	"afyonkarahisar": {Lat: 38.76376, Lon: 30.54034},
	"ağrı":           {Lat: 39.7216667, Lon: 43.0566667},
	"amasya":         {Lat: 40.65, Lon: 35.8333333},
	"ankara":         {Lat: 39.92077, Lon: 32.85411},
	"antalya":        {Lat: 36.88414, Lon: 30.70563},
	"artvin":         {Lat: 41.1833333, Lon: 41.8166667},
	"aydın":          {Lat: 37.8444, Lon: 27.8458},
	"balıkesir":      {Lat: 39.648369, Lon: 27.88261},
	"bilecik":        {Lat: 40.150131, Lon: 29.983061},
	"bingöl":         {Lat: 38.885349, Lon: 40.498291},
	"bitlis":         {Lat: 38.4, Lon: 42.1166667},
	"bolu":           {Lat: 40.739479, Lon: 31.611561},
	"burdur":         {Lat: 37.726909, Lon: 30.288876},
	"bursa":          {Lat: 40.18257, Lon: 29.06687},
	"çanakkale":      {Lat: 40.155312, Lon: 26.41416},
	"çankırı":        {Lat: 40.6, Lon: 33.6166667},
	"çorum":          {Lat: 40.5505556, Lon: 34.9555556},
	"denizli":        {Lat: 37.77652, Lon: 29.08639},
	"diyarbakır":     {Lat: 37.91441, Lon: 40.230629},
	"edirne":         {Lat: 41.6666667, Lon: 26.5666667},
	"elazığ":         {Lat: 38.680969, Lon: 39.226398},
	"erzincan":       {Lat: 39.75, Lon: 39.5},
	"erzurum":        {Lat: 39.9043189, Lon: 41.2678853},
	"eskişehir":      {Lat: 39.784302, Lon: 30.51922},
	"gaziantep":      {Lat: 37.06622, Lon: 37.38332},
	"giresun":        {Lat: 40.912811, Lon: 38.38953},
	"gümüşhane":      {Lat: 40.4602778, Lon: 39.4813889},
	"hakkari":        {Lat: 37.5833333, Lon: 43.7333333},
	"hatay":          {Lat: 36.4018488, Lon: 36.3498097},
	"ısparta":        {Lat: 37.7666667, Lon: 30.55},
	"mersin":         {Lat: 36.8, Lon: 34.6333333},
	"istanbul":       {Lat: 41.00527, Lon: 28.97696},
	"izmir":          {Lat: 38.41885, Lon: 27.12872},
	"kars":           {Lat: 40.59267, Lon: 43.077831},
	"kastamonu":      {Lat: 41.38871, Lon: 33.78273},
	"kayseri":        {Lat: 38.7333333, Lon: 35.4833333},
	"kırklareli":     {Lat: 41.7333333, Lon: 27.2166667},
	"kırşehir":       {Lat: 39.15, Lon: 34.1666667},
	"kocaeli":        {Lat: 40.8532704, Lon: 29.8815203},
	"konya":          {Lat: 37.8666667, Lon: 32.4833333},
	"kütahya":        {Lat: 39.4166667, Lon: 29.9833333},
	"malatya":        {Lat: 38.35519, Lon: 38.30946},
	"manisa":         {Lat: 38.619099, Lon: 27.428921},
	"kahramanmaraş":  {Lat: 37.5833333, Lon: 36.9333333},
	"mardin":         {Lat: 37.3122361, Lon: 40.735112},
	"muğla":          {Lat: 37.2152778, Lon: 28.3636111},
	"muş":            {Lat: 38.7432926, Lon: 41.5064823},
	"nevşehir":       {Lat: 38.62442, Lon: 34.723969},
	"niğde":          {Lat: 37.9666667, Lon: 34.6833333},
	"ordu":           {Lat: 40.9833333, Lon: 37.8833333},
	"rize":           {Lat: 41.02005, Lon: 40.523449},
	"sakarya":        {Lat: 40.7568793, Lon: 30.378138},
	"samsun":         {Lat: 41.292782, Lon: 36.33128},
	"siirt":          {Lat: 37.94429, Lon: 41.93288},
	"sinop":          {Lat: 42.0264222, Lon: 35.1550745},
	"sivas":          {Lat: 39.747662, Lon: 37.017879},
	"tekirdağ":       {Lat: 40.9833333, Lon: 27.5166667},
	"tokat":          {Lat: 40.3166667, Lon: 36.55},
	"trabzon":        {Lat: 41, Lon: 39.7333333},
	"tunceli":        {Lat: 39.1079868, Lon: 39.5401672},
	"şanlıurfa":      {Lat: 37.15, Lon: 38.8},
	"uşak":           {Lat: 38.682301, Lon: 29.40819},
	"van":            {Lat: 38.4941667, Lon: 43.38},
	"yozgat":         {Lat: 39.82, Lon: 34.8044444},
	"zonguldak":      {Lat: 41.456409, Lon: 31.798731},
	"aksaray":        {Lat: 38.36869, Lon: 34.03698},
	"bayburt":        {Lat: 40.255169, Lon: 40.22488},
	"karaman":        {Lat: 37.17593, Lon: 33.228748},
	"kırıkkale":      {Lat: 39.846821, Lon: 33.515251},
	"batman":         {Lat: 37.881168, Lon: 41.13509},
	"şırnak":         {Lat: 37.5163889, Lon: 42.4611111},
	"bartın":         {Lat: 41.6344444, Lon: 32.3375},
	"ardahan":        {Lat: 41.110481, Lon: 42.702171},
	"ığdır":          {Lat: 39.9166667, Lon: 44.0333333},
	"yalova":         {Lat: 40.65, Lon: 29.2666667},
	"karabük":        {Lat: 41.2, Lon: 32.6333333},
	"kilis":          {Lat: 36.718399, Lon: 37.12122},
	"osmaniye":       {Lat: 37.06805, Lon: 36.261589},
	"düzce":          {Lat: 40.843849, Lon: 31.15654},
}

func GetCityCoordinates(city string) (CityCoordinates, bool) {
	coords, ok := CityCoordinatesByName[city]
	return coords, ok
}

func LocationUsableForImpossibleTravel(location string) bool {
	c := strings.ToLower(strings.TrimSpace(location))
	if c == "" {
		return false
	}
	_, ok := GetCityCoordinates(c)
	return ok
}
