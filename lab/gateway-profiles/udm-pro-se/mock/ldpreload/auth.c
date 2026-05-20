#include "ubnthal_redirect.h"

/*
 * Lab-local root account compatibility.
 *
 * UDAPI checks the local root password during setup. In the VM path this uses a
 * fixed lab hash so the firmware can complete that check without copying real
 * appliance secrets into the repository or guest image.
 */
static const char lab_root_hash[] =
    "$6$unifilab$jsm7Qo3dSoGLzqmz60nSEke0ROYn/QQM5SYi3QUWS9x/dncS9bjfurjrP0X8hsQRCznof3lF3KcZ2SSloydhb.";


static struct passwd *lab_root_passwd(void) {
    static struct passwd pwd;

    pwd.pw_name = "root";
    pwd.pw_passwd = "x";
    pwd.pw_uid = 0;
    pwd.pw_gid = 0;
    pwd.pw_gecos = "root";
    pwd.pw_dir = "/root";
    pwd.pw_shell = "/bin/bash";
    return &pwd;
}

static struct spwd *lab_root_shadow(void) {
    static struct spwd sp;

    sp.sp_namp = "root";
    sp.sp_pwdp = (char *)lab_root_hash;
    sp.sp_lstchg = 20510;
    sp.sp_min = 0;
    sp.sp_max = 99999;
    sp.sp_warn = 7;
    sp.sp_inact = -1;
    sp.sp_expire = -1;
    sp.sp_flag = (unsigned long)-1;
    return &sp;
}

struct passwd *getpwnam(const char *name) {
    static struct passwd *(*real_getpwnam)(const char *) = NULL;
    if (real_getpwnam == NULL) {
        real_getpwnam = dlsym(RTLD_NEXT, "getpwnam");
    }

    if (udapi_user_check_enabled() && strcmp(name, "root") == 0) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: mocked getpwnam(\"root\")\n");
        }
        return lab_root_passwd();
    }

    return real_getpwnam != NULL ? real_getpwnam(name) : NULL;
}

struct spwd *getspnam(const char *name) {
    static struct spwd *(*real_getspnam)(const char *) = NULL;
    if (real_getspnam == NULL) {
        real_getspnam = dlsym(RTLD_NEXT, "getspnam");
    }

    if (udapi_user_check_enabled() && strcmp(name, "root") == 0) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: mocked getspnam(\"root\")\n");
        }
        return lab_root_shadow();
    }

    return real_getspnam != NULL ? real_getspnam(name) : NULL;
}

char *crypt(const char *key, const char *salt) {
    static char *(*real_crypt)(const char *, const char *) = NULL;
    if (real_crypt == NULL) {
        real_crypt = dlsym(RTLD_NEXT, "crypt");
    }

    if (real_crypt != NULL) {
        char *result = real_crypt(key, salt);
        if (result != NULL) {
            return result;
        }
    }

    if (udapi_user_check_enabled() && strcmp(key, "ubnt") == 0 &&
        (strcmp(salt, lab_root_hash) == 0 || strncmp(salt, "$6$unifilab$", 12) == 0)) {
        if (debug_enabled()) {
            fprintf(stderr, "ubnthal_redirect: mocked crypt(\"ubnt\", lab root salt)\n");
        }
        return (char *)lab_root_hash;
    }

    errno = EINVAL;
    return NULL;
}
