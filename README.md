Check_ssh_auth_methods is a Nagios/Icinga (or compatible) plugin to check available authentication methods for SSH.

---

<!-- TOC -->

- [What does this do](#what-does-this-do)
- [Installation](#installation)
    - [Download](#download)
    - [Building](#building)
- [Usage](#usage)
    - [Examples](#examples)
    - [Configuration](#configuration)
        - [Icinga](#icinga)

<!-- /TOC -->

---

# 1. What does this do

This plugin allows you to check if certain authentication methods are enabled or disabled.  
A common use case is to check if password authentication is disabled and publicKey authentication is enabled.

For this the plugin needs to connects to a given host and then evaluates the provided authentication methods.

# 2. Installation

Download the pre-build binary or build it yourself.  
Copy it to your existing plugins folder and add it to your monitoring software.  
Ensure that you have the OpenSSH client installed on the system (If `ssh -V` outputs OpenSSH on the system all is good)  
and the plugin is executable (`chmod 750`) and runs (`./check_ssh_auth_methods -h`).

## 2.1. Download

Download the latest release from [Releases](https://github.com/Massl123/check_ssh_auth_methods/releases).

## 2.2. Building

Ensure you have a working golang installation on your system and execute these commands:

~~~bash
mkdir -p ~/go/src/github.com/massl123/
cd ~/go/src/github.com/massl123/
git checkout https://github.com/Massl123/check_ssh_auth_methods.git
cd check_ssh_auth_methods
go build
~~~

You should now get the binary check_ssh_auth_methods

# 3. Usage

~~~config
allow:  authentication method must be allowed
forbid: authentication method must not be allowed (default if not stated otherwise)
ignore: authentication method is not checked
Arguments:
  -gssapikeyex value
      GssapiKeyex authentication, set to a[llow], f[orbid], i[gnore] (default ignored)
  -gssapiwithmic value
      GssapiWithMic authentication, set to a[llow], f[orbid], i[gnore] (default ignored)
  -host string
      Host to connect to (required)
  -hostbased value
      Hostbased authentication, set to a[llow], f[orbid], i[gnore]
  -keyboardinteractive value
      KeyboardInteractive authentication, set to a[llow], f[orbid], i[gnore]
  -none value
      None authentication, set to a[llow], f[orbid], i[gnore]
  -p string
      SSH port (default "22")
  -password value
      Password authentication, set to a[llow], f[orbid], i[gnore]
  -publickey value
      PublicKey authentication, set to a[llow], f[orbid], i[gnore] (default allowed)
  -t string
      SSH timeout in seconds (default "10")
  -u value
      SSH users to check, repeat argument for multiple users (default: root)
~~~

## 3.1. Examples

Check if none, password and hostbased authentication are disabled, publickey authentication enabled - ignore gssapi authentications

~~~bash
./check_ssh_auth_methods -host <host>
~~~

Checking multiple users (because authentication methods can be set per user in SSHD)

~~~bash
./check_ssh_auth_methods -host <host> -u root -u admin -u user1
~~~

Check if password authentication for user root is disabled - ignore all other values

~~~bash
./check_ssh_auth_methods -host <host> -password f -none i -hostbased i -keyboardinteractive i -publickey i
~~~

Example output

~~~text
CRITICAL for user(s) root, admin
CRITICAL:  root (None: ok, Hostbased: ok, Password: allowed but should be forbidden, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
CRITICAL: admin (None: ok, Hostbased: ok, Password: allowed but should be forbidden, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
~~~

~~~text
OK, checked user(s) root, admin, pi, user1, user2
OK:  root (None: ok, Hostbased: ok, Password: ok, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
OK: admin (None: ok, Hostbased: ok, Password: ok, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
OK:    pi (None: ok, Hostbased: ok, Password: ok, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
OK: user1 (None: ok, Hostbased: ok, Password: ok, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
OK: user2 (None: ok, Hostbased: ok, Password: ok, KeyboardInteractive: ok, PublicKey: ok, GssapiKeyex: ignore, GssapiWithMic: ignore)
~~~

## 3.2. Configuration

### 3.2.1. Icinga

CheckCommand defintion for Icinga

~~~icinga
object CheckCommand "ssh_auth_methods" {
    import "plugin-check-command"
    command = [ PluginDir + "/check_ssh_auth_methods" ]
    arguments += {
        "-gssapikeyex" = {
            description = "Allow gssapiKeyex authentication, set to a[llow], f[orbid], i[gnore] (default \"i\")"
            value = "$check_ssh_auth_methods_gssapikeyex$"
        }
        "-gssapiwithmic" = {
            description = "Allow gssapiWithMic authentication, set to a[llow], f[orbid], i[gnore] (default \"i\")"
            value = "$check_ssh_auth_methods_gssapiwithmic$"
        }
        "-host" = {
            description = "Host to connect to"
            required = true
            value = "$check_ssh_auth_methods_host$"
        }
        "-hostbased" = {
            description = "Allow hostbased authentication, set to a[llow], f[orbid], i[gnore] (default \"d\")"
            required = false
            value = "$check_ssh_auth_methods_hostbased$"
        }
        "-keyboardinteractive" = {
            description = "Allow keyboardInteractive authentication, set to a[llow], f[orbid], i[gnore] (default \"d\")"
            value = "$check_ssh_auth_methods_keyboardinteractive$"
        }
        "-none" = {
            description = "Allow none authentication, set to a[llow], f[orbid], i[gnore] (default \"d\")"
            required = false
            value = "$check_ssh_auth_methods_none$"
        }
        "-p" = {
            description = "SSH port (default \"22\")"
            required = false
            value = "$check_ssh_auth_methods_port$"
        }
        "-password" = {
            description = "Allow password authentication, set to a[llow], f[orbid], i[gnore] (default \"d\")"
            required = false
            value = "$check_ssh_auth_methods_password$"
        }
        "-publickey" = {
            description = "Allow publickey authentication, set to a[llow], f[orbid], i[gnore] (default \"a\")"
            value = "$check_ssh_auth_methods_publickey$"
        }
        "-t" = {
            description = "SSH timeout in seconds (default \"10\")"
            required = false
            value = "$check_ssh_auth_methods_timeout$"
        }
        "-u" = {
            description = "SSH users to check"
            repeat_key = true
            required = false
            value = "$check_ssh_auth_methods_users$"
        }
    }
    vars.check_ssh_auth_methods_gssapikeyex = "i"
    vars.check_ssh_auth_methods_gssapiwithmic = "i"
    vars.check_ssh_auth_methods_host = "$host.address$"
    vars.check_ssh_auth_methods_hostbased = "f"
    vars.check_ssh_auth_methods_keyboardinteractive = "f"
    vars.check_ssh_auth_methods_none = "f"
    vars.check_ssh_auth_methods_password = "f"
    vars.check_ssh_auth_methods_port = 22
    vars.check_ssh_auth_methods_publickey = "a"
    vars.check_ssh_auth_methods_timeout = 10
    vars.check_ssh_auth_methods_users = [ "root" ]
}
~~~
