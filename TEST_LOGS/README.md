To run this test, first, compile everything:

```
sudo snap install microk8s --classic

# clone juju
# clone pebble

go mod edit -replace github.com/canonical/pebble=<local_path>
DEBUG_JUJU=1 make build
DEBUG_JUJU=1 make install
# Install docker-buildx, add user ubuntu to docker group and 
DEBUG_JUJU=1 JUJU_BUILD_NUMBER=1 make operator-image
DEBUG_JUJU=1 JUJU_BUILD_NUMBER=1 make microk8s-operator-update
```

Deploy the env:
```
DEBUG_JUJU=1 ./_build/linux_amd64/bin/juju bootstrap test-k8s
./_build/linux_amd64/bin/juju add-model test

# Deploy apps
juju deploy mysql-k8s -n 5
juju deploy cos-lite --trust  # creates more pressure in the env
```

Now, capture logs for each container:

```
for i in {0..4}; do kubectl logs -n test mysql-k8s-$i -c mysql -f  > log.$i & done
```

Then, cause it to fail:
```
# Create load
juju deploy mysql-test-app app
juju relate app:database mysql-k8s
juju run mysql-test-app/leader start-continuous-writes

# Wait until everything settles
juju scale-application mysql-k8s 7
```

Now, containers will start to error

Check the log with:
```
cat <log-file> | grep "\/v1\/health\|daemon.go ServerHTTP\|api_health.go v1Health" | grep <remote client port number>
```

That will give the key entries for a given client (liveness probe in this case) connection.
