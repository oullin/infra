# Supervisor Quick-Start for GoLand Users

This guide provides the specific steps to create a Supervisor configuration file that avoids the false syntax errors caused by the Nginx plugin in IDEs like GoLand.

The solution is to use a clean configuration file with no comments (`#` or `;`), as the plugin may misinterpret them.

## 1. Create the Configuration File

Create the `.conf` file for your service in Supervisor's configuration directory.

```bash
sudo nano /etc/supervisor/conf.d/oullin-api.conf
```

## 2. Add the Comment-Free Configuration

Paste the following configuration into the file. This version is functionally correct for Supervisor and will not trigger errors in the IDE's Nginx parser.

```ini
[program:oullin-api]
command=/usr/bin/docker compose up
directory=/home/gocanto/Sites/oullin/api
user=gocanto
autostart=true
autorestart=true
startsecs=10
startretries=3
stopsignal=INT
stopwaitsecs=30
killasgroup=true
stdout_logfile=/var/log/supervisor/oullin-api.log
stderr_logfile=/var/log/supervisor/oullin-api.err.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
```

## 3. Activate the Service

After saving the file, use `supervisorctl` to load and start your service.

**Step 1: Reread Configurations**

This tells Supervisor to check for the new file.

```bash
sudo supervisorctl reread
```
*Expected Output:* `oullin-api: available`

**Step 2: Apply the Changes**

This command loads and starts the new service.

```bash
sudo supervisorctl update
```
*Expected Output:* `oullin-api: added process group`

## 4. Verify the Service is Running

Check the status to confirm that Supervisor has successfully started your program.

```bash
sudo supervisorctl status
```
*Expected Output:* `oullin-api             RUNNING   pid 1234, uptime 0:00:15`

Your service is now running, and your IDE should no longer show incorrect syntax errors for this file.
