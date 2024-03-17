# vtyang
Yang based VTY

```
mkdir -p /tmp/vtyang/run
echo '{"users": {"user": [{"name": "hiroki"}]}}' > /tmp/config.json
./vtyang agent --dbpath /tmp/config.json --run-path /tmp/vtyang/run
```
