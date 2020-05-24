# logserverd
<p>GOlang daemon for Linux to access Apache or NGINX log files from a browser</p>

# Introduction
<p>For websites I use PHP as my primary development language. As I won't always have direct access to my sites from protected environments I've developed a way to do everything browser based. The logserverd project is writte in GOlang and is meant to serve as a deamon on Linux for easy logfiles access either directly or via a reverse proxy.</p>

# Coding and debugging challenges
<p>For coding I use ICEcoder (check out https://icecoder.net). I've yet to encounter an environment where this tool does not work ...</p>
The next challenge was to find a clear and secure way to access the webserver log files in the /var/log path. After some tricks (remounting /var/log/httpd for example) I decided to go for a more structural and more clean option which is why and where I ended up in GOlang.

# Operations dependencies
<ul>
<li>For deamonizing I'm using a library from https://github.com/takama/deamon</li>
<li>Logs are in a single logging directory</li>
<li>Log name convention: <em>sitename-type.log</em> (site = example.com, type = <em>error</em> or <em>access</em></li>
</ul>

# Features
<ul>
<li>Last X lines of the logfile (defaults to 25 lines)</li>
<li>Auto refresh (default is off)</li>
<li>Reverse date/time (default is oldest to newerst)</li>
<li>Filter based on our client IP address (default is to show all entries)</li>
<li>Filter /_log and /_<em>icecoder alias</em></li>
</ul>

# Usage notes
<ul>
<li>Compile the code "go build logserverd.go"</li>
<li>Install the deamon to the sbin path with "sudo cp logserverd /usr/local/bin/sbin/logserverd"</li>
<li>Install the daemon into services with "sudo logserverd install"</li>
<li>Start the deamon with "sudo logserverd start"</li>
</ul>
<p>The service is now ready and waiting on port 7000</p>

# Apache log server configuration (recommended setup)
<p>...</p>

# Notes
<p>I'm still developing (security related) and documenting the code. Updates to follow in the next weeks. If you want to use the current code, look for configurable items in the code before compiling/using.</p>

# Security notes
<p><b>Use this deamon at your own risk! I'm still improving the code to prevent script attacks etc.</b></p>
<p>To ensure you are using all services safely, apply Let's Encrypt (or similar) to the website and enforce HTTPS.</p>
<p>For safety (and to prevent a lot of uninvited guests peeking in your logs) I strongly recommend to apply Basic Authentication (or similar) to the /_log (or your own configured path).</p>
<p><b>Full documentation on INSTALL is pending.</b></p>
