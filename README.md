### What is this project
This project is forked from Masterminds/glide, if you wish to read the official readme, please see [here](https://github.com/Masterminds/glide)

### Why not dep
The glide offical has given a suggestion that migrating from glide to dep(which is a similar tool published by Go community), but as a Chinese for some you know reason, glide's mirror is very usefull, and dep is not support this function yet.

### Why fork the project but not commit code to offical
Glide project is be in a state of support rather than active feature development, so I have to fork the project and develop some features that I wanted.

### What is the difference between offical glide and this project
- fix the bug which move command failed on windows
- mirrors support base url and local mirrors.yaml. see [mirror](https://github.com/fatcat22/glide/blob/master/docs/commands.md#glide-mirror) document for detail.
