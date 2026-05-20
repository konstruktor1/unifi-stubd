/^[[:space:]]*return 301 https:\/\/\$host\$request_uri;/ {
    print "    # UDM Pro SE lab HTTP preview."
    print "    # Docker maps host 127.0.0.1:9080 to container port 80. A"
    print "    # stock redirect points the browser at host port 443, which"
    print "    # is not the mapped HTTPS port in this lab."
    print "    include /usr/share/unifi-core/http/errors.conf;"
    print "    include /usr/share/unifi-core/http/shared-server-defaults.conf;"
    print "    include /usr/share/unifi-core/http/shared-post-setup-server.conf;"
    next
}
{ print }
