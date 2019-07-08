import subprocess

db_PTH_CMD = "--dbpath  "
db_path = "\"C:/Program Files/MongoDB/Server/4.0/data\""
cmd = "mongod"

cmd += " {} {}".format(db_PTH_CMD, db_path)

proc = subprocess.Popen(cmd)

input()
print("Closing server")
proc.kill()
