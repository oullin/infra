# Install autossh

```shell
brew install autossh
```
```shell
which autossh
```
> We’ll need that full path in the plist below (on an M1/M2 Mac it’ll often be /opt/homebrew/bin/autossh; on Intel,
> /usr/local/bin/autossh).

# Create/Copy the LaunchAgent plist
> Create a file at ~/Library/LaunchAgents/com.oullin.ssh-tunnel.plist with these contents:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <!-- Unique label -->
  <key>Label</key>
  <string>com.gocanto.ssh-tunnel</string>

  <!-- Command & args -->
  <key>ProgramArguments</key>
  <array>
    <string>/FULL/PATH/TO/autossh</string>
    <string>-M</string><string>0</string>
    <string>-N</string>
    <string>-L</string><string>15432:localhost:5432</string>
    <string>-o</string><string>ServerAliveInterval=60</string>
    <string>-o</string><string>ServerAliveCountMax=3</string>
    <string>-i</string><string>/Users/YOUR_MAC_USERNAME/.ssh/id_rsa</string>
    <string>gocanto@YOUR_VPS_IP</string>
  </array>

  <!-- Start at login -->
  <key>RunAtLoad</key>
  <true/>

  <!-- Restart if it ever quits -->
  <key>KeepAlive</key>
  <true/>

  <!-- Logs (optional) -->
  <key>StandardOutPath</key>
  <string>/Users/YOUR_MAC_USERNAME/Library/Logs/ssh-tunnel.out.log</string>
  <key>StandardErrorPath</key>
  <string>/Users/YOUR_MAC_USERNAME/Library/Logs/ssh-tunnel.err.log</string>
</dict>
</plist>
```
- Replace /FULL/PATH/TO/autossh with what which autossh returned.
- Replace YOUR_MAC_USERNAME with your macOS user.
- Replace YOUR_VPS_IP with your VPS’s public IP.

# Load (and start) the agent
Check it’s running
```shell
launchctl list | grep com.gocanto.ssh-tunnel
```
View logs
```shell
tail -f ~/Library/Logs/ssh-tunnel.err.log
```
Stop/unload
```shell
launchctl unload ~/Library/LaunchAgents/com.oullin.ssh-tunnel.plist
```

# Validate the plist
```shell
plutil -lint ~/Library/LaunchAgents/com.oullin.ssh-tunnel.plist
```
If you see any errors, fix them in your XML (missing tags, stray characters, etc.) before proceeding.

# Verify it’s running
```shell
launchctl list | grep com.oullin.ssh-tunnel
```
You should see your label listed.
