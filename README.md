# Purpose
A sample app used for testing how your environment handles idle mysql sessions

# How does it work
Loops forever executing a MYSQL PING request and running query `show full processlist` at an interval defined by `$INTERVAL`

# Update the manifest

The environmental parameters will override the settings in VCAP_SERVICES.  You will have to set the mysql service instance name under services

```
env:
  #HOSTNAME: 127.0.0.1
  #SQLUSERNAME: user
  #SQLPASSWORD: password
  #DATABASE: dbname
  #INTERVAL: 315
services:
  - danlsql
```