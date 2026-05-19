#!/bin/bash
# CC statusLine script — token usage in status bar.

STATS_FILE="${XDG_CACHE_HOME:-$HOME/.cache}/cc-focus/stats.json"

python3 -c "
import json
try:
    with open('$STATS_FILE') as f:
        stats = json.load(f)
    if not stats:
        print('\U0001F3AF cc-focus | no data')
        exit()
    s = max(stats.values(), key=lambda x: x.get('total_turns', 0))
    sid = s['session_id'][:8]
    inp, out, cached = s['input_tokens'], s['output_tokens'], s['cache_tokens']
    turns, model = s['total_turns'], s['last_model']
    cost = inp*0.14/1e6 + out*1.10/1e6 + cached*0.014/1e6
    print(f'\U0001F3AF {sid} | ↩{turns} | ↓{inp} ↑{out} ⇗{cached} | \${cost:.4f} | {model}')
except:
    print('\U0001F3AF cc-focus')
" 2>/dev/null
