// Package shared provides the entrypoint to Sourcegraph's single docker
// image. It has functionality to setup the shared environment variables, as
// well as create the Procfile for goreman to run.
package shared

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/server/internal/goreman"
)

// FrontendInternalHost is the value of SRC_FRONTEND_INTERNAL.
const FrontendInternalHost = "127.0.0.1:3090"

// defaultEnv is environment variables that will be set if not already set.
var defaultEnv = map[string]string{
	// Sourcegraph services running in this container
	"SRC_GIT_SERVERS":       "127.0.0.1:3178",
	"SEARCHER_URL":          "http://127.0.0.1:3181",
	"REPO_UPDATER_URL":      "http://127.0.0.1:3182",
	"QUERY_RUNNER_URL":      "http://127.0.0.1:3183",
	"SRC_SYNTECT_SERVER":    "http://127.0.0.1:9238",
	"SYMBOLS_URL":           "http://127.0.0.1:3184",
	"REPLACER_URL":          "http://127.0.0.1:3185",
	"LSIF_SERVER_URL":       "http://127.0.0.1:3186",
	"SRC_HTTP_ADDR":         ":8080",
	"SRC_HTTPS_ADDR":        ":8443",
	"SRC_FRONTEND_INTERNAL": FrontendInternalHost,
	"GITHUB_BASE_URL":       "http://127.0.0.1:3180", // points to github-proxy

	// Limit our cache size to 100GB, same as prod. We should probably update
	// searcher/symbols to ensure this value isn't larger than the volume for
	// CACHE_DIR.
	"SEARCHER_CACHE_SIZE_MB": "50000",
	"REPLACER_CACHE_SIZE_MB": "50000",
	"SYMBOLS_CACHE_SIZE_MB":  "50000",

	// Used to differentiate between deployments on dev, Docker, and Kubernetes.
	"DEPLOY_TYPE": "docker-container",

	// enables the debug proxy (/-/debug)
	"SRC_PROF_HTTP": "",

	"LOGO":          "t",
	"SRC_LOG_LEVEL": "warn",

	// TODO other bits
	// * DEBUG LOG_REQUESTS https://github.com/sourcegraph/sourcegraph/issues/8458
}

// Set verbosity based on simple interpretation of env var to avoid external dependencies (such as
// on github.com/sourcegraph/sourcegraph/pkg/env).
var verbose = os.Getenv("SRC_LOG_LEVEL") == "dbug" || os.Getenv("SRC_LOG_LEVEL") == "info"

// Main is the main server command function which is shared between Sourcegraph
// server's open-source and enterprise variant.
func Main() {
	flag.Parse()
	log.SetFlags(0)

	// Ensure CONFIG_DIR and DATA_DIR

	// Load $CONFIG_DIR/env before we set any defaults
	{
		configDir := SetDefaultEnv("CONFIG_DIR", "/etc/sourcegraph")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			log.Fatalf("failed to ensure CONFIG_DIR exists: %s", err)
		}

		err = godotenv.Load(filepath.Join(configDir, "env"))
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("failed to load %s: %s", filepath.Join(configDir, "env"), err)
		}

		// Load the legacy config file if it exists.
		//
		// TODO(slimsag): Remove this code in the next significant version of
		// Sourcegraph after 3.0.
		configPath := os.Getenv("SOURCEGRAPH_CONFIG_FILE")
		if configPath == "" {
			configPath = filepath.Join(configDir, "sourcegraph-config.json")
		}
		_, err = os.Stat(configPath)
		if err == nil {
			if err := os.Setenv("SOURCEGRAPH_CONFIG_FILE", configPath); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Next persistence
	{
		SetDefaultEnv("SRC_REPOS_DIR", filepath.Join(DataDir, "repos"))
		SetDefaultEnv("LSIF_STORAGE_ROOT", filepath.Join(DataDir, "lsif-storage"))
		SetDefaultEnv("CACHE_DIR", filepath.Join(DataDir, "cache"))
	}

	// Special case some convenience environment variables
	if redis, ok := os.LookupEnv("REDIS"); ok {
		SetDefaultEnv("REDIS_ENDPOINT", redis)
	}

	data, err := json.MarshalIndent(SrcProfServices, "", "  ")
	if err != nil {
		log.Println("Failed to marshal default SRC_PROF_SERVICES")
	} else {
		SetDefaultEnv("SRC_PROF_SERVICES", string(data))
	}

	for k, v := range defaultEnv {
		SetDefaultEnv(k, v)
	}

	// Now we put things in the right place on the FS
	if err := copySSH(); err != nil {
		// TODO There are likely several cases where we don't need SSH
		// working, we shouldn't prevent setup in those cases. The main one
		// that comes to mind is an ORIGIN_MAP which creates https clone URLs.
		log.Println("Failed to setup SSH authorization:", err)
		log.Fatal("SSH authorization required for cloning from your codehost. Please see README.")
	}
	if err := copyNetrc(); err != nil {
		log.Fatal("Failed to copy netrc:", err)
	}

	// TODO validate known_hosts contains all code hosts in config.

	nginx, err := nginxProcFile()
	if err != nil {
		log.Fatal("Failed to setup nginx:", err)
	}

	procfile := []string{
		nginx,
		`frontend: env CONFIGURATION_MODE=server frontend`,
		`gitserver: gitserver`,
		`query-runner: query-runner`,
		`symbols: symbols`,
		`lsif-server: node /lsif-server.js | grep -v 'Listening for HTTP requests'`,
		`management-console: management-console`,
		`searcher: searcher`,
		`github-proxy: github-proxy`,
		`repo-updater: repo-updater`,
		`syntect_server: sh -c 'env QUIET=true ROCKET_ENV=production ROCKET_PORT=9238 ROCKET_LIMITS='"'"'{json=10485760}'"'"' ROCKET_SECRET_KEY='"'"'SeerutKeyIsI7releuantAndknvsuZPluaseIgnorYA='"'"' ROCKET_KEEP_ALIVE=0 ROCKET_ADDRESS='"'"'"127.0.0.1"'"'"' syntect_server | grep -v "Rocket has launched" | grep -v "Warning: environment is"' | grep -v 'Configured for production'`,
	}
	procfile = append(procfile, ProcfileAdditions...)

	redisStoreLine, err := maybeRedisStoreProcFile()
	if err != nil {
		log.Fatal(err)
	}
	if redisStoreLine != "" {
		procfile = append(procfile, redisStoreLine)
	}
	redisCacheLine, err := maybeRedisCacheProcFile()
	if err != nil {
		log.Fatal(err)
	}
	if redisCacheLine != "" {
		procfile = append(procfile, redisCacheLine)
	}
	if redisStoreLine != "" && redisCacheLine != "" {
		// We boot up Redis (as opposed to connecting to an addr in a
		// user-supplied ENV var), which means we can migrate it
		go func() {
			time.Sleep(30 * time.Second) // Give Redis time to start up
			err := migrateRedisInstances()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	if line, err := maybePostgresProcFile(); err != nil {
		log.Fatal(err)
	} else if line != "" {
		procfile = append(procfile, line)
	}

	procfile = append(procfile, maybeZoektProcFile()...)

	const goremanAddr = "127.0.0.1:5005"
	if err := os.Setenv("GOREMAN_RPC_ADDR", goremanAddr); err != nil {
		log.Fatal(err)
	}

	err = goreman.Start(goremanAddr, []byte(strings.Join(procfile, "\n")))
	if err != nil {
		log.Fatal(err)
	}
}

func migrateRedisInstances() error {
	// 1. Check whether redispool.Cache and redispool.Store are connected to
	// different instances
	storeAddr, ok := os.LookupEnv("REDIS_STORE_ENDPOINT")
	if !ok {
		return errors.New("could not connect to redis-store: REDIS_STORE_ENDPOINT not set")
	}

	storeConn, err := redis.Dial("tcp", storeAddr)
	if err != nil {
		return err
	}

	cacheAddr, ok := os.LookupEnv("REDIS_CACHE_ENDPOINT")
	if !ok {
		return errors.New("could not connect to redis-cache: REDIS_CACHE_ENDPOINT not set")
	}

	cacheConn, err := redis.Dial("tcp", cacheAddr)
	if err != nil {
		return err
	}

	storeRunID, err := getRunID(storeConn)
	if err != nil {
		return err
	}
	cacheRunID, err := getRunID(cacheConn)
	if err != nil {
		return err
	}

	if cacheRunID == storeRunID {
		// Nothing to migrate
		return nil
	}

	// 2. Delete the keys on both instances that are now outdated
	const rcacheDataVersionToDelete = "v1"

	err = deleteKeysWithPrefix(storeConn, rcacheDataVersionToDelete)
	if err != nil {
		return err
	}

	err = deleteKeysWithPrefix(cacheConn, rcacheDataVersionToDelete)
	if err != nil {
		return err
	}

	return nil
}

func getRunID(c redis.Conn) (string, error) {
	infos, err := redis.String(c.Do("INFO", "server"))
	if err != nil {
		return "", err
	}

	for _, l := range strings.Split(infos, "\n") {
		if strings.HasPrefix(l, "run_id:") {
			s := strings.Split(l, ":")
			return s[1], nil
		}
	}
	return "", errors.New("no run_id found")
}

func deleteKeysWithPrefix(c redis.Conn, prefix string) error {
	pattern := prefix + ":*"

	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving keys with pattern %q", pattern)
		}

		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return err
		}

		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return err
		}
		keys = append(keys, k...)
		if iter == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	const batchSize = 1000
	var batch = make([]interface{}, batchSize, batchSize)

	for i := 0; i < len(keys); i += batchSize {
		j := i + batchSize
		if j > len(keys) {
			j = len(keys)
		}
		currentBatchSize := j - i

		for bi, v := range keys[i:j] {
			batch[bi] = v
		}

		// We ignore whether the number of deleted keys matches what we have in
		// `batch`, because in the time since we constructed `keys` some of the
		// keys might have expired
		_, err := c.Do("DEL", batch[:currentBatchSize]...)
		if err != nil {
			return fmt.Errorf("failed to delete keys: %s", err)
		}
	}

	return nil
}
