# drpc-provider-estimator

Provider Estimator is a capacity measurement tool for estimating the number of providers needed for a given workload. 

## Usage

To use Provider Estimator, simply run the following command:
```
drpc-provider-estimator [OPTIONS]
```
## Options

```
Application Options:
  -t, --target=            target host
  -c, --chain=             chain id (default: 100)
  -d, --step-duration=     step duration in minutes (default: 3)
  -s, --source=            source eth host (default: https://eth.drpc.org)
  -e, --stop-on-events     stop on events
  -o, --csv-output=        csv output file
  -m, --mode=              mode. Can be spam or prepared (default: spam)
  -p, --spam-profile=      spam profile
  -r, --prepared-requests= prepared requests folder
  -u, --prepared-cu=       prepared request cu cost (default: 0)
  -l, --request-label=     request label for dshackle
  -a, --load-levels=       load levels
  -i, --insecure           certificate
  ```

## run.sh

The repository includes a pre-configured script for running several standard workloads on a Dshackle instance. To use this, you need to compile the necessary tools, disable TLS on dshackle and execute the following command:
 ```bash
 ./run.sh <host>:<port>
 ```

## Contributing

If you would like to contribute to Provider Estimator, please fork the repository and submit a pull request.

## License

Provider Estimator is licensed under the MIT License. See LICENSE for more information.