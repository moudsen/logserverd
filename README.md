# logserverd
GOlang daemon for Linux to access Apache or NGINX log files from a browser

# Introduction
For websites I use PHP as my primary development language. As I won't always have direct access to my sites from protected environments I've developed a way to do everything browser based. The logserverd project is writte in GOlang and is meant to serve as a deamon on Linux for easy logfiles access either directly or via a reverse proxy.

# My coding and debugging challenges
For coding on the server I use <em>ICEcoder</em> (check out https://icecoder.net), a browser based code editor. I've yet to encounter an environment where this tool does not work ...!

The next challenge was to find a clear and secure way to access the webserver log files in the /var/log path. After some tricks (remounting /var/log/httpd for example) I decided to go for a more structural and more clean option which is in my case I decided to use GOlang.

# Operations/dependencies
- Commandline access (including root privileges)
- GO compiler (recent version)
- For deamonizing I'm using a library from https://github.com/takama/deamon
- Apache or NGINX logs are in a single logging directory
- Log name convention: <em>sitename-type.log</em> (site = example.com, type = <em>error</em> or <em>access</em>
- Apache or NGINX reverse proxy modules (to serve as standard HTTPS)

# Features
- Access or Error logs (refer to name convention); ```?type=<access|error>```
- Last X lines of the logfile (defaults to 25 lines); ```&lines=<number>```
- Auto refresh (default is off); ```&refresh=<seconds>```
- Reverse date/time (default is oldest to newerst); ```&reverse=<0|1>```
- Filter based on our client IP address (default is to show all entries); ```&filter="<0|1>```
- Filter ```/_log``` and ```/_icecoder-alias```
- Reverse proxy detection (correct source IP address for filtering the log file)

# Usage/installation notes
- Install the go library ```go get github.com/takama/deamon```
- Compile the code ```go build logserverd.go```
- Copy the deamon to the sbin path with ```sudo cp logserverd /usr/local/bin/sbin/logserverd```
- Install and enable the daemon into standard services with ```sudo logserverd install```
- Start the deamon with ```sudo logserverd start```

The service is now ready and waiting on port 7000 ...

# Apache log server configuration (recommended setup)
Use reverse proxy for Apache and protect the ```/_log directory``` (add to virtual host config or as general config file in case your wish to provide the service to multiple sites)

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

# Notes
I'm still developing (security related) and documenting the code. Updates to follow in the next weeks. If you want to use the current code, look for configurable items in the code before compiling/using.

# Next steps/to-do
- More configuraton of parameters/behaviour (port from command-line, ```/var/log/path/subdirectory```, fixate site)
- Configuration file (/etc/logserverd.conf)
- Enhance security (enforce allowed paths only, non-root daemon)

# Security notes
<b>Use this deamon at your own risk! I'm still improving the code to prevent script attacks etc.</b>

To ensure you are using all services safely, it is strongly recommended to apply Let's Encrypt (or similar) to the (whole) website and to enforce HTTPS in all cases. It's free and easy to setup!

For safety (and to prevent a lot of uninvited guests peeking in your logs) I strongly recommend to apply Basic Authentication (or similar) to the ```/_log``` (or your own configured path).
