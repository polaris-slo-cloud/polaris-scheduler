package collections

import "sync"

var (
	_ EntityLocker = (*EntityLockerImpl)(nil)
	_ EntityLock   = (*entityLockImpl)(nil)
)

// Default implementation of the EntityLocker.
type EntityLockerImpl struct {
	// Thread-safe map that stores pointers to entityLockImpl objects.
	activeLocks *sync.Map
}

// Callback function passed to an entityLockImpl to allow the lock to delete itself from the map.
type deleteLockFn func()

type entityLockImpl struct {
	// The name of the entity that is covered by this lock.
	name string

	// Used to tell the parent locker to delete this lock.
	deleteLock deleteLockFn

	// The number of lock requests that have been queued for this object.
	// When the lock is currently in use, this has a value of 1.
	// If additional requests are waiting, the count is greater than 1.
	// When the count hits 0, the is marked as deleted and removed from the map.
	// The mutex needs to be locked before accessing this.
	queuedRequests int

	// If true, this lock has already been deleted and should not be used any more.
	// The mutex needs to be locked before accessing this.
	deleted bool

	// The mutex for queuedRequests, deleted, and the cond variable.
	mutex sync.Mutex

	// The condition variable used for waiting on the lock.
	cond sync.Cond
}

func NewEntityLockerImpl() *EntityLockerImpl {
	el := &EntityLockerImpl{
		activeLocks: &sync.Map{},
	}
	return el
}

// Creates a new entityLockImpl with queuedRequests = 0 and the mutex locked.
func newEntityLockImpl(entityName string, deleteLock deleteLockFn) *entityLockImpl {
	lock := &entityLockImpl{
		name:           entityName,
		deleteLock:     deleteLock,
		queuedRequests: 0,
		deleted:        false,
		mutex:          sync.Mutex{},
	}
	lock.cond = *sync.NewCond(&lock.mutex)
	lock.mutex.Lock()
	return lock
}

func (locker *EntityLockerImpl) Lock(name string) EntityLock {
	lock := locker.getOrCreateLock(name)
	lock.queuedRequests++
	defer lock.mutex.Unlock()

	// We need to wait on this lock only if we are not the first one in the queue.
	if lock.queuedRequests > 1 {
		lock.cond.Wait()
	}
	return lock
}

// Gets an existing lock or creates a new one for the specified entity name.
// In either case, the mutex of the returned lock is in a locked state.
func (locker *EntityLockerImpl) getOrCreateLock(name string) *entityLockImpl {
	var lock *entityLockImpl

	// Check if there is an existing lock object and, if so, lock it and ensure that it has not been deleted.
	if untypedLock, ok := locker.activeLocks.Load(name); ok {
		lock = untypedLock.(*entityLockImpl)
		lock.mutex.Lock()
		if lock.deleted {
			lock.mutex.Unlock()
			lock = nil
		}
	}

	if lock == nil {
		newLock := newEntityLockImpl(name, func() {
			locker.activeLocks.Delete(name)
		})

		// Loop until we manage to store either our new lock or we obtain a non-deleted existing lock.
		for lock == nil {
			if untypedLock, loaded := locker.activeLocks.LoadOrStore(name, newLock); loaded {
				// If another goroutine has stored a lock in the meantime, lock it and ensure that it has not been deleted.
				lock = untypedLock.(*entityLockImpl)
				lock.mutex.Lock()
				if lock.deleted {
					lock.mutex.Unlock()
					lock = nil
				}
			}
		}
	}

	return lock
}

func (lock *entityLockImpl) Name() string {
	return lock.name
}

func (lock *entityLockImpl) Unlock() {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	lock.queuedRequests--
	if lock.queuedRequests > 0 {
		// If there are more goroutines waiting for this lock, wake up the next one.
		lock.cond.Signal()
	} else {
		lock.deleted = true
		lock.deleteLock()
	}
}
