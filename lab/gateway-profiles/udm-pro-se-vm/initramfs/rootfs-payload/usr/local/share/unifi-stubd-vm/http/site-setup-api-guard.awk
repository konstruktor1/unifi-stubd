# Insert lab-only guards before the public UniFi OS API block.
# These endpoints affect the host appliance state on real hardware. In the VM
# lab they stay explicit and reviewable instead of being driven by the UI.
/^[[:space:]]*# UniFi OS public API/ && !inserted {
    print "    # unifi-stubd VM lab API guard."
    print "    # Keep destructive host-affecting setup actions explicit in this QEMU lab."
    print "    location = /api/system/reboot { return 403; }"
    print "    location = /api/system/reset { return 403; }"
    print "    location = /api/setup/reset { return 403; }"
    print ""
    inserted = 1
}
{ print }
