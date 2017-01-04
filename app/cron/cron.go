// This library implements a cron spec parser and runner.  See the README for
// more details.
package cron

import (
	"github.com/astaxie/beego"
	"sort"
	"time"
)

// 删除job的检查函数，返回true则删除
type RemoveCheckCheckFunc func(e *Entry) bool

type RemoveCheckFunc struct {
	RemoveFunc func(e *Entry) bool
}

//var fun RemoveCheckFunc

var drop = RemoveCheckFunc{
	RemoveFunc: func(e *Entry) bool {
		if e.Status {
			return true
		}
		return false
	},
}

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Cron struct {
	entries  []*Entry
	stop     chan struct{}
	add      chan *Entry
	remove   chan RemoveCheckFunc
	snapshot chan []*Entry
	running  bool
}

// Job is an interface for submitted cron jobs.
type Job interface {
	Run()
}

// The Schedule describes a job's duty cycle.
type Schedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// The schedule on which this job should be run.
	Schedule Schedule

	// The next time the job will run. This is the zero time if Cron has not been
	// started or this entry's schedule is unsatisfiable
	Next time.Time

	// The last time this job was run. This is the zero time if the job has never
	// been run.
	Prev time.Time

	// The Job to run.
	Job Job

	// 判断改entry是否被执行了
	Status bool
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

// New returns a new Cron job runner.
func New() *Cron {
	return &Cron{
		entries:  nil,
		add:      make(chan *Entry),
		remove:   make(chan RemoveCheckFunc, 1),
		stop:     make(chan struct{}),
		snapshot: make(chan []*Entry),
		running:  false,
	}
}

// A wrapper that turns a func() into a cron.Job
type FuncJob func()

func (f FuncJob) Run() { f() }

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddFunc(spec string, cmd func()) error {
	return c.AddJob(spec, FuncJob(cmd))
}

// AddFunc adds a Job to the Cron to be run on the given schedule.
func (c *Cron) AddJob(spec string, cmd Job) error {
	schedule, err := Parse(spec)
	if err != nil {
		return err
	}
	c.Schedule(schedule, cmd)
	return nil
}

func (c *Cron) RemoveJob(cbf RemoveCheckCheckFunc) {
	cb := RemoveCheckFunc{
		RemoveFunc: cbf,
	}
	c.remove <- cb
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *Cron) Schedule(schedule Schedule, cmd Job) {
	entry := &Entry{
		Schedule: schedule,
		Job:      cmd,
		Status:   false,
	}
	if !c.running {
		c.entries = append(c.entries, entry)
		return
	}
	beego.Critical("entry is", entry)
	c.add <- entry
}

// Entries returns a snapshot of the cron entries.
func (c *Cron) Entries() []*Entry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

// Start the cron scheduler in its own go-routine.
func (c *Cron) Start() {
	c.running = true
	go c.run()
}

// Run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run() {
	beego.Info("start run")
	// Figure out the next activation times for each entry.
	now := time.Now().Local()

	beego.Info("c.entry is", c.entries)
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
		beego.Critical("entry is", entry.Next)
	}

	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))

		var effective time.Time
		beego.Emergency(len(c.entries))
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			//beego.Critical("1111111111")
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			effective = now.AddDate(10, 0, 0)
		} else if c.entries[0].Status {
			//beego.Critical("22222222222")
			//c.remove <- drop
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = c.entries[0].Next
			//beego.Critical("this next end")
		}
		beego.Critical("effective is", effective)

		select {
		case now = <-time.After(effective.Sub(now)):
			//beego.Critical("come to now")
			// Run every entry whose next time was this effective time.
			for _, e := range c.entries {
				beego.Critical(now, e.Status)
				if e.Next != effective || e.Status {
					//beego.Critical("break run this entry!")
					break
				}
				//beego.Critical("start to run!")
				go e.Job.Run()
				e.Prev = e.Next
				e.Status = true
				//e.Next = e.Schedule. Next(effective)
				e.Next = e.Prev
				//beego.Critical("drop")
			}
			/*
				select {
				case c.remove <- drop:
					beego.Critical("xiexie")
				case cb := <-c.remove:
					beego.Critical("drop one")
					newEntries := make([]*Entry, 0)
					for _, e := range c.entries {
						beego.Critical("now e is", e)
						if !cb.RemoveFunc(e) {
							beego.Critical("drop oneone")
							newEntries = append(newEntries, e)
						}
					}
					c.entries = newEntries
				}
			*/
			c.remove <- drop
			continue

		case newEntry := <-c.add:
			beego.Critical("come to add")
			c.entries = append(c.entries, newEntry)

			newEntry.Next = newEntry.Schedule.Next(now)

		case cb := <-c.remove:
			beego.Critical("drop one")
			newEntries := make([]*Entry, 0)
			for _, e := range c.entries {
				/*
					if !cb(e) {
						newEntries = append(newEntries, e)
					}
				*/
				if !cb.RemoveFunc(e) {
					beego.Critical("drop oneone")
					newEntries = append(newEntries, e)
				}
			}
			c.entries = newEntries

		case <-c.snapshot:
			c.snapshot <- c.entrySnapshot()

		case <-c.stop:
			return
		}

		// 'now' should be updated after newEntry and snapshot cases.
		now = time.Now().Local()
	}
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
func (c *Cron) Stop() {
	if !c.running {
		return
	}
	c.stop <- struct{}{}
	c.running = false
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *Cron) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Schedule: e.Schedule,
			Next:     e.Next,
			Prev:     e.Prev,
			Job:      e.Job,
		})
	}
	return entries
}
