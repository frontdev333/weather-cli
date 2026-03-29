package ui

import (
	"fmt"
	"frontdev333/weather-cli/internal/domain"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	cyan   = "\x1b[36m"
	blue   = "\x1b[34m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	gray   = "\x1b[90m"
)

func useColor() bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

func wrap(text, color string) string {
	if !useColor() {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, reset)
}

func colorTemp(celsius float64, text string) string {
	if celsius < 0 {
		return wrap(text, blue)
	}

	if celsius <= 25 {
		return wrap(text, yellow)
	}

	return wrap(text, red)
}

func f642ClrdStr(num float64) string {
	res := strconv.FormatFloat(num, 'f', -1, 64)
	return colorTemp(num, res)
}

func iconForCondition(cond string) string {
	cond = strings.ToLower(cond)
	switch {
	case strings.Contains(cond, "гроза") && strings.Contains(cond, "град"):
		return "⛈️🧊"
	case strings.Contains(cond, "гроза"):
		return "⛈️"
	case strings.Contains(cond, "туман"):
		return "🌫️"
	case strings.Contains(cond, "морось"):
		return "🌦️"
	case strings.Contains(cond, "дожд"):
		return "🌧️"
	case strings.Contains(cond, "снег"):
		return "❄️"
	case strings.Contains(cond, "ясно"):
		return "☀️"
	case strings.Contains(cond, "облач") || strings.Contains(cond, "пасмур"):
		return "⛅️"
	default:
		return "❓"
	}
}

func Header(city string, cached bool, fetchedAt time.Time) string {
	var res strings.Builder

	res.WriteString(city)
	res.WriteString(fmt.Sprintf(" • обновлено %.0f мин назад", time.Since(fetchedAt).Minutes()))
	if cached {
		res.WriteString(" • из кэша")
	}
	res.WriteString("\n")
	return res.String()
}

func RenderToday(t domain.Today) string {
	builder := &strings.Builder{}

	tb := tabwriter.NewWriter(builder, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tb, "Сегодня в %s [%s]\n", wrap(t.City, bold), iconForCondition(t.Conditions))
	fmt.Fprintf(tb, "Температура:\t%s (ощущается как %s°C)\n", f642ClrdStr(t.Temperature), f642ClrdStr(t.FeelsLike))
	fmt.Fprintf(tb, "Условие:\t%s\n", t.Conditions)
	fmt.Fprintf(tb, "Ветер:\t%.1f м/с (180°)\n", t.WindSpeed)
	fmt.Fprintf(tb, "Влажность:\t%d%%\n", t.Humidity)
	fmt.Fprintf(tb, "Давление:\t%.1f hPa\n", t.Pressure)
	fmt.Fprintf(tb, "Видимость:\t%.1f\n", t.Visibility)
	fmt.Fprintf(tb, "Осадки (1ч):\t%.1f мм\n", t.Precipitation)
	tb.Flush()

	return builder.String()
}

func RenderMenu() string {
	var res strings.Builder
	res.WriteString(wrap("────────────────────────────────────────────────────────────\n", gray))
	res.WriteString(wrap("[1] Почасовой (12 ч)  [2] На 7 дней  [C] Сменить город  [R] Обновить  [Q] Выход\n", gray))
	return res.String()
}

func RenderHourly(list []domain.HourlyEntry) string {
	var res strings.Builder
	tb := tabwriter.NewWriter(&res, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tb, "Почасовой прогноз (%d часов):\n", len(list))
	fmt.Fprintf(tb, "Время\t| t°C\t| Осадки\t| Ветер м/с\n")
	fmt.Fprintf(tb, wrap("------------------------------------\n", gray))

	for _, v := range list {
		fmt.Fprintf(tb, " %s\t| %s\t| %.0f%%\t| %.1f\n", v.Time.Format("15:04"), f642ClrdStr(v.Temperature), v.PrecipitationProbability, v.WindSpeed)
	}

	tb.Flush()

	return res.String()
}

func RenderDaily(list []domain.DailyEntry) string {
	var res strings.Builder
	tb := tabwriter.NewWriter(&res, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tb, "Прогноз на неделю:\n")
	fmt.Fprintf(tb, "Дата\t| Мин°C\t| Макс°C\t| Осадки\n")
	fmt.Fprintf(tb, wrap("------------------------------------\n", gray))
	for _, v := range list {
		fmt.Fprintf(tb, "%s %s\t| %s \t| %s \t| %.1f\n", v.Date.Format("02 Jan"), iconForCondition(v.Conditions), f642ClrdStr(v.MinTemperature), colorTemp(v.MaxTemperature, f642ClrdStr(v.MaxTemperature)), v.PrecipitationProbability)
	}

	tb.Flush()
	return res.String()
}
