/^[[:space:]]*# UniFi OS public API/ && !inserted {
    print "    # UDM Pro SE lab reset guard."
    print "    # These endpoints would normally reboot or reset the console."
    print "    # The firmware lab blocks them because host-mutating actions are no-ops here."
    print "    location = /api/setup/reset { return 403; }"
    print "    location = /api/system/reboot { return 403; }"
    print "    location = /api/system/reset { return 403; }"
    inserted = 1
}
{ print }
