/**
 * Store Exports
 * Central export for all store-related modules
 */

// Entity Store
export {
  useEntityStore,
  User,
  isStale,
  createFreshMetadata,
  setLoadingMetadata,
  setSuccessMetadata,
  setErrorMetadata,
  TTL_CONFIG,
  type CacheMetadata,
  type CachedEntity,
  type SceneState,
  type EventState,
  type UserState,
  type EntityStoreState,
  type SceneActions,
  type EventActions,
  type UserActions,
  type EntityStoreActions,
  type EntityStore,
} from './entityStore';

// Auth Store
export { authStore, useAuth, type User as AuthUser } from './authStore';
