use near_lake_framework::{LakeConfig, LakeConfigBuilder};
use std::env;
use std::mem::MaybeUninit;
use std::sync::{Mutex, MutexGuard};
use once_cell::sync::Lazy;
use crate::pusher::redis::RedisPusher;

pub const INDEXER: &str = "map-near-indexer-s3";
pub const REDIS: &str = "redis";
pub static PROJECT_CONFIG: Lazy<Env> = Lazy::new(init_env_config);
static mut REDIS_PUSHER: MaybeUninit<Mutex<RedisPusher>> = MaybeUninit::uninit();
static BLOCK_HEIGHT: &str = "block_height";

pub struct Env {
    pub(crate) start_block_height_from_cache: bool,
    pub(crate) start_block_height: u64,
    pub(crate) redis_url: String,
    pub(crate) pub_list: String,
    pub(crate) mcs: String,
    pub(crate) test: bool,
    pub(crate) log_file: String,
    pub(crate) log_level: String,
}

pub async fn init_lake_config() -> LakeConfig {
    let mut current_height = PROJECT_CONFIG.start_block_height;
    if PROJECT_CONFIG.start_block_height_from_cache {
        current_height = get_synced_block_height().await + 1;
    }

    tracing::info!(target: INDEXER, "start stream from block {}", current_height);
    if PROJECT_CONFIG.test {
        LakeConfigBuilder::default()
            .testnet()
            .start_block_height(current_height)
            .build()
            .expect("failed to start block height")
    } else {
        LakeConfigBuilder::default()
            .mainnet()
            .start_block_height(current_height)
            .build()
            .expect("failed to start block height")
    }
}

pub fn init_env_config() -> Env {
    for (key, value) in env::vars() {
        println!("{}: {}", key, value);
    }
    Env {
        start_block_height_from_cache: env::var("START_BLOCK_HEIGHT_FROM_CACHE")
            .unwrap()
            .parse::<bool>()
            .unwrap(),
        start_block_height: env::var("START_BLOCK_HEIGHT")
            .unwrap()
            .parse::<u64>()
            .unwrap(),
        redis_url: env::var("REDIS_URL").unwrap(),
        pub_list: env::var("PUB_LIST").unwrap(),
        mcs: env::var("MCS").unwrap(),
        test: env::var("TEST")
            .unwrap()
            .parse::<bool>()
            .unwrap(),
        log_file: env::var("LOG_FILE").unwrap(),
        log_level: env::var("LOG_LEVEL").unwrap(),
    }
}

pub async fn init_redis_pusher() {
    // Make it
    let pusher = RedisPusher::new(&PROJECT_CONFIG.redis_url, &PROJECT_CONFIG.pub_list)
        .await.expect("New redis pusher fail");
    // Store it to the static var, i.e. initialize it
    unsafe {
        REDIS_PUSHER.write(Mutex::new(pusher));
    }
}

pub fn redis_publisher() -> MutexGuard<'static, RedisPusher> {
    unsafe {

        // Now we give out a shared reference to the data, which is safe to use
        // concurrently.
        REDIS_PUSHER.assume_init_ref().lock().unwrap()
    }
}

pub async fn get_synced_block_height() -> u64 {
    let value = redis_publisher().get(BLOCK_HEIGHT).await;
    if value.is_some() {
        let height: u64 = serde_json::from_str(value.unwrap().as_str()).unwrap();
        height
    } else { 0 }
}

pub async fn update_synced_block_height(height: u64) {
    redis_publisher().set(BLOCK_HEIGHT, serde_json::to_string(&height).unwrap()).await;
}
