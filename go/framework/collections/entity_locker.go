package collections

// Allows obtaining and waiting for locks on entities that do not support locking themselves.
//
// The entities are identified using their names and the actual objects are not needed for this locker.
// This is useful, e.g., when a node in a cluster needs to be locked - the node can be locked by its name
// before fetching the actual object.
type EntityLocker interface {
	// Obtains a lock for the entity with the specified name.
	// If the entity is already locked, the goroutine will block until the lock becomes available.
	Lock(name string) EntityLock
}

// Represents ownership of a lock on an entity obtained from an EntityLocker.
type EntityLock interface {
	// The name of the entity that is covered by this lock.
	Name() string

	// Unlocks the entity.
	Unlock()
}
