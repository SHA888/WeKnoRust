pub mod tenant;

use sqlx::{Pool, Postgres};
use thiserror::Error;
use dotenvy::dotenv;
use std::env;

#[derive(Debug, Error)]
pub enum RepoError {
    #[error("database url not configured")] 
    MissingDatabaseUrl,
    #[error(transparent)]
    Sqlx(#[from] sqlx::Error),
}

pub type PgPool = Pool<Postgres>;

// Initialize a Postgres pool from DATABASE_URL
pub async fn init_pool() -> Result<PgPool, RepoError> {
    dotenv().ok();
    let url = env::var("DATABASE_URL").map_err(|_| RepoError::MissingDatabaseUrl)?;
    let pool = sqlx::postgres::PgPoolOptions::new()
        .max_connections(10)
        .connect(&url)
        .await?;
    Ok(pool)
}

// Initialize a Postgres pool from a provided URL
pub async fn init_pool_from(url: &str) -> Result<PgPool, RepoError> {
    let pool = sqlx::postgres::PgPoolOptions::new()
        .max_connections(10)
        .connect(url)
        .await?;
    Ok(pool)
}
