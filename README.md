# upsya
UDESA Python evaluator proof of concept.

### Usage
1. Run project with PYTHON3 environment variable set.
    ```sh
    PYTHON3=python3 go run . -evalglob=testdata/evaluations/*/*.py
    ```
    This will run uncontainerized code. Very unsafe for production. See below on how to containerize.

2. Navigate to address as logged to stdout.

### Containerization
Download [`soypat/gontainer`](https://github.com/soypat/gontainer) and create
a filesystem for use. The instructions over there are straightforward and simple.
I've found the size of a python installation is usually around 320MB (without libraries)
so I'd suggest creating a VFS with at least 500MB. The Alpine **Mini-Root filesystem** is enough to run python.

Once you have a mounted filsystem with a [python install on it](https://github.com/soypat/gontainer/blob/master/Gockerfile) run the following

```sh
sudo su # You must be root to modify containerized filesystem
PYTHON3=python3 GONTAINER_FS=/mnt/your-vfs go run . -evalglob=testdata/evaluations/*/*.py
```
Ready. All code run from there on out will be containerized.