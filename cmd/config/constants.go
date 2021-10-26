/*
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

// Config value separator
const (
	ValueSeparator = ","
)

// Top level common ENVs
const (
	EnvAccessKey    = "MINIO_ACCESS_KEY"
	EnvSecretKey    = "MINIO_SECRET_KEY"
	EnvAccessKeyOld = "MINIO_ACCESS_KEY_OLD"
	EnvSecretKeyOld = "MINIO_SECRET_KEY_OLD"
	EnvBrowser      = "MINIO_BROWSER"
	EnvDomain       = "MINIO_DOMAIN"
	EnvRegionName   = "MINIO_REGION_NAME"
	EnvPublicIPs    = "MINIO_PUBLIC_IPS"
	EnvFSOSync      = "MINIO_FS_OSYNC"
	EnvFSODirect    = "MINIO_FS_ODIRECT"
	EnvArgs         = "MINIO_ARGS"
	EnvDNSWebhook   = "MINIO_DNS_WEBHOOK_ENDPOINT"

	EnvUpdate = "MINIO_UPDATE"

	EnvEndpoints = "MINIO_ENDPOINTS" // legacy
	EnvWorm      = "MINIO_WORM"      // legacy
	EnvRegion    = "MINIO_REGION"    // legacy

	EnvIAMRefreshSecInterval = "MINIO_IAM_REFRESH_SEC_INTERVAL"
	EnvSTSMinDuration = "MINIO_STS_MIN_DURATION"
	EnvDefaultFilesystemPath = "DEFAULT_FILESYSTEM_PATH"
	EnvWekaAccessSecret = "WEKA_ACCESS_SECRET"

	EnvCrawlSleepPerFolder = "MINIO_CRAWL_SLEEP_PER_FOLDER"
	EnvCrawlStartDelay = "MINIO_CRAWL_START_DELAY"
	EnvDataUsageUpdateDirCycles = "MINIO_DATA_USAGE_UPDATE_DIR_CYCLES"
	EnvDataUsageSleepPerFile = "MINIO_DATA_USAGE_SLEEP_PER_FILE"

	EnvStaleUploadsExpiry = "MINIO_STALE_UPLOADS_EXPIRY"
	EnvStaleUploadsCleanupInterval = "MINIO_STALE_UPLOADS_CLEANUP_INTERVAL"

	EnvETCDOnly = "MINIO_ETCD_ONLY"

	EnvMaxBucketsLimit = "MAX_BUCKETS_LIMIT"
)
