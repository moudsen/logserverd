# LOGSERVERD
GOlang daemon for Linux to access Apache or NGINX log files from a browser.

# Introduction
For websites I use PHP as my primary development language. As I won't always have direct access to my sites (specifically) from protected environments I've developed a way to do everything browser based. The logserverd project is meant to serve as a deamon on any Linux VPS for easy logfiles access either directly (or via a reverse proxy).

## My coding and debugging challenges
For coding on the server I use <em>ICEcoder</em> (check out https://icecoder.net), a browser based code editor. I've yet to encounter an environment where this tool does not work ...!

The next challenge after phpMyAdmin and other tools besides the coding was to find an easy and secure way to access the webserver log files in the ```/var/log path```. After some tricks (remounting ```/var/log/httpd``` for example) I decided to go for a more structural and more clean option for which in my case I decided to use GOlang (multi platform/OS flavors, single binary).

# Prerequisites
- Commandline access (including root privileges to install the daemon, but also depends on access rights to ```/var/log```: as long as read-only right is granted the daemon will work)
- GO compiler (recent version)
- For deamonizing the program I'm using a very nice and easy to use library from https://github.com/takama/deamon
- Apache or NGINX logs are in a single logging directory (considering multi-path support at a later stage)
- Log name convention: <em>sitename-type.log</em> (site = example.com, type = <em>error</em> or <em>access</em> (considering to make this configurable while guarding security level)
- Apache or NGINX reverse proxy modules (to serve as standard HTTPS and to make the deamon part of the main site(s))

# Features
- Access or Error logs (refer to name convention); ```?type=<access|error>```
- Read logfiles from other sites; ```&site=<server name>``` (*)
- Last X lines of the logfile (defaults to 25 lines); ```&lines=<number>```
- Auto refresh (default is off); ```&refresh=<seconds>```
- Reverse date/time (default is oldest to newerst); ```&reverse=<0|1>```
- Filter based on our client IP address (default is to show all entries); ```&filter="<0|1>```
- Filter ```/_log``` and ```/_icecoder-alias``` (_coder)
- Reverse proxy detection (correct source IP address for filtering the log file)

(*) Site name characters may only contain a-z,0-9 or .-_ characters

# Usage/installation notes
- Install the go library ```go get github.com/takama/deamon```
- Compile the code ```go build logserverd.go```
- Copy the deamon to the sbin path with ```sudo cp logserverd /usr/local/bin/sbin/logserverd```
- Install and enable the daemon into standard services with ```sudo logserverd install```
- Start the deamon with ```sudo logserverd start```

The service is now ready and waiting on port 7000 ...

The deamon also listens to stop or uninstall:
- Stop the deamon with ```sudo logserverd stop```
- Uninstall the deamon with ```sudo logserverd uninstall```

# Apache log server configuration (recommended and secure setup)
There are numerous examples on the internet how to setup access control for Apache and NGINX. I am using the following configuration for Apache to somehow secure the logging server access, but feel free to adapt according to your needs.

Configure the reverse proxy for Apache and protect the ```/_log directory``` (add configuration to any virtual host config or as general config file in case your wish to provide the service to multiple sites)

```
  <Location /_log>
    ProxyPass http://127.0.0.1:7000/_log
    ProxyPassReverse http://127.0.0.1:7000/_log
    AuthType Basic
    AuthName "Authentication Required"
    AuthUserFile "/var/pathto/filename.passwd"
    Require valid-user
    Order allow,deny
    Allow from all
</Location>
```

I'm assuming basic knowledge how to configure Apache and to create a passwd file. Hint: ```htpasswd -c /pathto/filename.passwd myusername```.

# Notes
I'm still developing (security related) and documenting the code. If you want to use the current code, look for configurable items in the code before compiling/using (like the exclusion of _log and _coder log lines).

# To-do
- More configuraton of parameters/behaviour (port from command-line, ```/var/log/path/subdirectory```, fixate site)
- Configuration file (/etc/logserverd.conf)
- Enhance security (enforce allowed paths only, non-root daemon)

# Security notes
To ensure you are using all services safely, it is strongly recommended to install certificates with Let's Encrypt (or similar) to the (whole) website and to enforce HTTPS in all cases. It's free and easy to setup! (and prevents your username/password/coding, perhaps including sensitive information, going over the internet unencrypted ...)

For safety (and to prevent a lot of uninvited guests peeking in your logs) I strongly recommend to apply Basic Authentication (or similar) to the ```/_log``` (or your own configured path).

# Feedback
Feel free to open an issue or to contact me directly at mark.oudsen@puzzl.nl (please put "[LOGSERVERD]" in the subject).
