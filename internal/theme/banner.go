// internal/theme/banner.go
package theme

import (
	"fmt"
	"math/rand"
)

var quotes = []string{
	`"The net is vast and infinite." — Ghost in the Shell`,
	`"I never asked for this." — Adam Jensen`,
	`"Shall we play a game?" — WOPR`,
	`"In a world of locked doors, the man with the key is king."`,
	`"I'm in." — Every 90s hacker movie`,
	`"The more things change, the more they stay the same."`,
}

func (t *Theme) Banner(version string) string {
	if !t.IsRetro() {
		return ""
	}

	quote := quotes[rand.Intn(len(quotes))]

	banner := fmt.Sprintf(`
    %s
    %s
    %s
    %s  %s
    %s  %s
    %s  %s
    %s   %s
    %s   %s
    %s
    %s  %s
    %s
    %s
`,
		Green.Render("╔══════════════════════════════════════╗"),
		Green.Render("║                                      ║"),
		Green.Render("║       ▄▄▄                            ║"),
		Green.Render("║      █▀█▀█"), Green.Render("┏━╸╺┳╸╻                 ║"),
		Green.Render("║      █▄█▄█"), Green.Render("┃   ┃ ┃                 ║"),
		Green.Render("║      ╰█▀█╯"), Green.Render("┗━╸ ╹ ┗━╸              ║"),
		Green.Render("║       ╰─╯"), Green.Render("d a t a d o g           ║"),
		Green.Render("║          "), Green.Render("c o n t r o l           ║"),
		Green.Render("║                                      ║"),
		Green.Render("║  ["+version+"]"), Green.Render("   ◄◄ JACK IN ►►          ║"),
		Green.Render("║                                      ║"),
		Green.Render("╚══════════════════════════════════════╝"),
	)

	banner += "\n  " + DimGreen.Italic(true).Render(quote) + "\n"

	return banner
}
