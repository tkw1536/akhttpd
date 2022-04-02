<!doctype html>
<html lang="en">
<title>User {{.User}} - akhttpd - Authorized Keys HTTP Daemon</title>
<style>
    body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif;
        line-height: 1.5;
        color: #000;
        background: #fff;
    }

    a {
        color: #000;
        text-decoration: underline;
    }

    code {
        background: lightgray;
        padding: 5px;
    }

    code.replace, code.key {
        user-select: all;
    }

    code.block {
        margin: 10px;
    }
</style>
<p>
    This page contains a list of SSH Keys for the
    {{if eq (.Source) ("github") }}
        <a href="https://github.com/{{ .User }}" target="_blank" rel="noreferrer noopener">GitHub User {{.User}}</a>
    {{else}}
        <a>User {{.User}}</a>
    {{end}}. 
    This page is powered by <a href="/">akhttpd</a>.
</p>
<p>
    Click each entry to copy it to the clipboard.
</p>
<ul>
{{ range .Keys }}
<li><pre><code class="block key">{{.}}</code></pre></li>
{{end}}
</ul>

<p>
    To install these keys on an ssh server, you could do something like:
</p>
<p>
    <code class="block replace">
        curl -L localhost:8080/{{.User}} > .ssh/authorized_keys
    </code>
</p>
<p>
    For convenience, this service also exposes a script to do this automatically.
    Using this script will overwrite any existing SSH Keys for your user.
    You can use it like:
</p>
<p>
    <code class="block replace">
        curl -L localhost:8080/{{.User}}.sh | sh
    </code>
</p>

<script>
(function(keys) {
    var handleClick = function() {
        var text = this.innerText.trim();
        if (!navigator.clipboard) {
            prompt('Copy to Clipboard', text);
            return;
        }
        navigator.clipboard.writeText(text);
    }
    for (var i = 0; i < keys.length; i++ ) {
        keys[i].addEventListener('click', handleClick);
    }
})(document.getElementsByClassName('key'));
</script>
<script>
    (function (elements) {
        var update = function (element) {
            var originalText = element.innerHTML;
            var hostname = location.host;
            var host = location.protocol + "//" + hostname;
            element.innerHTML = originalText
                .replace('http://localhost:8080', host)
                .replace('localhost:8080', hostname);
        };
        for (var i = 0; i < elements.length; i++) {
            update(elements[i]);
        }
    })(document.getElementsByClassName('replace'));
</script>
