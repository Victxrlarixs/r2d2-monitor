package ui

// Reaction represents a visual ASCII art state and its associated dialogue lines.
type Reaction struct {
	Art      []string
	Dialogue []string
}

// R2Reactions maps robot moods to their respective ASCII art and dialogue pools.
var R2Reactions = map[string]Reaction{
	"idle": {
		Art: []string{
			`   ___________  `,
			`  /  ___ ___  \ `,
			` |  | (O) |  | |`,
			` |--+-----+--|-|`,
			` | [=]   [=]   |`,
			` | [ ]---[ ]   |`,
			` | [_________] |`,
			` |   |_____| | |`,
			` |___|     |___|`,
			` /____|___|____\`,
		},
		Dialogue: []string{
			"*Beep boop* Systems nominal.",
			"*Bloop* All threads stable.",
			"*Whistle* Monitoring active.",
			"*Bweep* Neural link stable.",
			"*Chirp* Standing by for input.",
			"*Bleep* Scanning background tasks.",
			"*Trill* Power levels within limits.",
			"*Whistle* Keeping an eye on things.",
		},
	},
	"thinking": {
		Art: []string{
			`   ___________  `,
			`  /  ___ ___  \ `,
			` |  | (?) |  | |`,
			` |--+-----+--|-|`,
			` | [!]   [!]   |`,
			` | [ ]---[ ]   |`,
			` | [=========] |`,
			` |   |_____| | |`,
			` |___|     |___|`,
			` /____|___|____\`,
		},
		Dialogue: []string{
			"*Bloop bloop* Analyzing...",
			"*Tshhh* Recalibrating sensors.",
			"*Beep?* Processing data...",
			"*Whirr* Calculating delta values.",
			"*Bweep* Accessing kernel telemetry.",
			"*Bloop* Optimizing refresh buffers.",
		},
	},
	"scanning": {
		Art: []string{
			`   ___________  `,
			`  /  ___ ___  \ `,
			` |  | [~] |  | |`,
			` |--+-----+--|-|`,
			` | [?]   [?]   |`,
			` | [ ]---[ ]   |`,
			` | [~~~~~~~~~] |`,
			` |   |_____| | |`,
			` |___|     |___|`,
			` /____|___|____\`,
		},
		Dialogue: []string{
			"*Scanner hum* Target locked!",
			"*Beeeep* Filtering results.",
			"*Zzzzt* Signal acquired.",
			"*Bleep* Matching PID signatures.",
			"*Whistle* Locating system anomalies.",
			"*Hummm* Scanning process tree.",
		},
	},
	"success": {
		Art: []string{
			`   ___________  `,
			`  /  ___ ___  \ `,
			` |  | (+) |  | |`,
			` |--+++++++--|-|`,
			` | [+]   [+]   |`,
			` | [ ]---[ ]   |`,
			` | [!!!!!!!!]  |`,
			` |   |_____| | |`,
			` |___|     |___|`,
			` /____|___|____\`,
		},
		Dialogue: []string{
			"*Happy whistle* Completed.",
			"*Joyful beeps* Optimized.",
			"*Whistle* Task accomplished.",
			"*Bleep bloop* Command successful.",
			"*Trill* Target eliminated.",
		},
	},
	"alarm": {
		Art: []string{
			`   ___________  `,
			`  /  _!_ _!_  \ `,
			` |  | [!] |  | |`,
			` |--+!!!!!!+--||`,
			` | [!]   [!]   |`,
			` | [ ]---[ ]   |`,
			` | [OVERLOAD!] |`,
			` |   |_____| | |`,
			` |___|     |__|`,
			` /____|___|____\`,
		},
		Dialogue: []string{
			"*SCREEEE* CPU CRITICAL!!",
			"*WHEEEEE* Thermal warning!",
			"*ALARM!* System overload!",
			"*Bweeeep* Resource exhaustion detected!",
			"*Siren* Kernel pressure warning!",
		},
	},
}
