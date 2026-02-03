package workerpool

import (
	"log"
	"net"
	"sync"
)

type Job func(net.Conn)

type WorkerPool struct {
	maxWorkers    int
	queueCapacity int // New: Capacity of the job queue
	jobQueue      chan net.Conn
	workerWg      sync.WaitGroup
	stop          chan struct{}
	hconn         Job
}

func NewWorkerPool(maxWorkers int, queueCapacity int, handler Job) *WorkerPool {
	p := &WorkerPool{
		maxWorkers:    maxWorkers,
		queueCapacity: queueCapacity,
		jobQueue:      make(chan net.Conn, queueCapacity), // Make jobQueue buffered
		stop:          make(chan struct{}),
		hconn:         handler,
	}

	for i := 0; i < maxWorkers; i++ {
		p.workerWg.Add(1)
		go p.worker(i + 1)
	}

	return p
}

func (p *WorkerPool) worker(id int) {
	defer p.workerWg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case conn := <-p.jobQueue:
			p.hconn(conn)
		case <-p.stop:
			log.Printf("Worker %d stopping", id)
			return
		}
	}
}

func (p *WorkerPool) Submit(conn net.Conn) {
	select {
	case p.jobQueue <- conn:
		// Job successfully submitted
	default:
		log.Println("Job queue full, rejecting connection from", conn.RemoteAddr())
		conn.Close() // Close connection immediately if job queue is full
	}
}

func (p *WorkerPool) Stop() {
	close(p.stop)
	p.workerWg.Wait()
	close(p.jobQueue)
	log.Println("Worker pool stopped")
}
