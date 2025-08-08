## 🎬 Demo

### Interactive TUI with Sorting & Filtering
![TUI Demo](assets/tui-demo-optimized.gif)

**Features shown:**
- Real-time port scanning visualization
- Sort results by port, state, service, or latency
- Filter by port state (open/closed/filtered)
- Keyboard navigation and shortcuts
- Progress tracking with ETA

### Command-Line Interface
![CLI Demo](assets/cli-demo-optimized.gif)

**Features shown:**
- Multiple scan profiles (quick, web, database)
- Custom port ranges and lists
- Banner grabbing for service detection
- JSON/CSV export formats
- Rate limiting and performance tuning

### Quick Start

```bash
# Quick scan with top 100 ports
portscan scan example.com --profile quick

# Scan specific ports with banner grabbing
portscan scan target.com --ports 80,443,8080 --banners

# Export results as JSON
portscan scan 192.168.1.1 --ports 1-1000 --output json > results.json

# High-performance scan
portscan scan 10.0.0.0/24 --workers 200 --rate 10000
```

### TUI Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `s` | Sort results |
| `f` | Filter results |
| `o` | Toggle open ports only |
| `r` | Reset all filters |
| `/` | Search in banners |
| `?` | Show help |
| `q` | Quit |
