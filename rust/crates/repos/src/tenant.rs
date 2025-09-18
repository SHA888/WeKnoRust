use crate::{PgPool, RepoError};
use sqlx::{self, Acquire, Postgres};
use wk_types::Tenant;

#[derive(Debug, Clone)]
pub struct TenantRepository<'a> {
    pub pool: &'a PgPool,
}

#[derive(sqlx::FromRow, Debug, Clone)]
struct TenantRow {
    pub id: i64,
    pub name: Option<String>,
    pub description: Option<String>,
    pub api_key: Option<String>,
    pub storage_used: Option<i64>,
}

impl<'a> TenantRepository<'a> {
    pub fn new(pool: &'a PgPool) -> Self { Self { pool } }

    pub async fn get_by_id(&self, id: i64) -> Result<Option<Tenant>, RepoError> {
        let row = sqlx::query_as::<_, TenantRow>(
            r#"SELECT id, name, description, api_key, storage_used FROM tenants WHERE id = $1"#,
        )
        .bind(id)
        .fetch_optional(self.pool)
        .await?;

        Ok(row.map(|r| Tenant {
            id: Some(r.id as u32),
            name: r.name,
            description: r.description,
            api_key: r.api_key,
            storage_used: r.storage_used,
        }))
    }

    pub async fn create(&self, t: &Tenant) -> Result<Tenant, RepoError> {
        let rec = sqlx::query_as::<_, TenantRow>(
            r#"
            INSERT INTO tenants (name, description, api_key, storage_used)
            VALUES ($1, $2, $3, COALESCE($4, 0))
            RETURNING id, name, description, api_key, storage_used
            "#,
        )
        .bind(t.name.as_deref())
        .bind(t.description.as_deref())
        .bind(t.api_key.as_deref())
        .bind(t.storage_used)
        .fetch_one(self.pool)
        .await?;

        Ok(Tenant {
            id: Some(rec.id as u32),
            name: rec.name,
            description: rec.description,
            api_key: rec.api_key,
            storage_used: rec.storage_used,
        })
    }

    pub async fn update(&self, t: &Tenant) -> Result<Tenant, RepoError> {
        let id = t.id.ok_or(RepoError::MissingDatabaseUrl)? as i64; // reuse error type for brevity
        let rec = sqlx::query_as::<_, TenantRow>(
            r#"
            UPDATE tenants
               SET name = COALESCE($2, name),
                   description = COALESCE($3, description),
                   api_key = COALESCE($4, api_key),
                   storage_used = COALESCE($5, storage_used)
             WHERE id = $1
         RETURNING id, name, description, api_key, storage_used
            "#,
        )
        .bind(id)
        .bind(t.name.as_deref())
        .bind(t.description.as_deref())
        .bind(t.api_key.as_deref())
        .bind(t.storage_used)
        .fetch_one(self.pool)
        .await?;

        Ok(Tenant {
            id: Some(rec.id as u32),
            name: rec.name,
            description: rec.description,
            api_key: rec.api_key,
            storage_used: rec.storage_used,
        })
    }

    pub async fn delete(&self, id: i64) -> Result<(), RepoError> {
        sqlx::query(r#"DELETE FROM tenants WHERE id = $1"#)
            .bind(id)
            .execute(self.pool)
            .await?;
        Ok(())
    }

    // AdjustStorageUsed with pessimistic lock, clamp to >= 0
    pub async fn adjust_storage_used(&self, tenant_id: i64, delta: i64) -> Result<(), RepoError> {
        let mut tx = self.pool.begin().await?;

        // SELECT ... FOR UPDATE to lock the row
        let mut used: i64 = sqlx::query_scalar::<_, i64>(
            r#"SELECT COALESCE(storage_used, 0) FROM tenants WHERE id = $1 FOR UPDATE"#,
        )
        .bind(tenant_id)
        .fetch_one(&mut *tx)
        .await?;

        used += delta;
        if used < 0 { used = 0; }

        sqlx::query(r#"UPDATE tenants SET storage_used = $2 WHERE id = $1"#)
            .bind(tenant_id)
            .bind(used)
            .execute(&mut *tx)
            .await?;

        tx.commit().await?;
        Ok(())
    }
}
