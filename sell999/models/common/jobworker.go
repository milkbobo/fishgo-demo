/*
 * Copyright (c) 2016 - hongbeibang Co.,Ltd. All rights reserved.
 *
 * @Author: jdlau
 * @Date: 2017-03-15 10:39:39
 * @Last Modified by: jdlau
 * @Last Modified time: 2017-03-15 13:45:13
 */

// 模拟工作／工作者模式
// 通过环境变量指定工作者数量、队列数量
package common

import (
	. "github.com/milkbobo/fishgoweb/web"
)

// 工作
type job struct {
	data    interface{}            // 工作要处理的内容
	handler func(data interface{}) // 工作处理方法
}

// 工作管道

var jobQueue chan job

// 工作者
type Worker struct {
	workerPool chan chan job // 保存工作管道的管道
	jobChannel chan job      // 工作管道
	quit       chan bool     // 停止管道
}

// 创建工作者
func NewWorker(workerPool chan chan job) Worker {
	return Worker{
		workerPool: workerPool,
		jobChannel: make(chan job),
		quit:       make(chan bool),
	}
}

// 启动工作者
func (w Worker) Start() {
	go func() {
		for {

			// 保存工作管道到工作池管道中
			w.workerPool <- w.jobChannel

			select {
			// 读取JobChannel的值
			case job := <-w.jobChannel:
				// 处理
				job.handler(job.data)
			// 停止
			case <-w.quit:
				return
			}

		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// 调度器
type Dispatcher struct {
	WorkerPool chan chan job // 工作管道池
}

// 创建调度器
func NewDispatcher(maxWorker int) *Dispatcher {
	return &Dispatcher{
		WorkerPool: make(chan chan job, maxWorker),
	}
}

// 运行调度器
func (d *Dispatcher) Run(maxWorker, maxQueue int) {
	jobQueue = make(chan job, maxQueue)

	// 启动指定数量的工作者
	for i := 0; i < maxWorker; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

// 调度处理
func (d *Dispatcher) dispatch() {
	for {
		select {
		case singleJob := <-jobQueue:
			// 接收到工作
			go func(singleJob job) {
				jobChannel := <-d.WorkerPool

				jobChannel <- singleJob
			}(singleJob)
		}
	}
}

// 添加工作
func CommonAddJob(data interface{}, handler func(data interface{})) {
	work := job{
		data:    data,
		handler: handler,
	}

	// 将工作推入工作管道
	jobQueue <- work
}

func init() {
	InitDaemon(func() {
		// 启动job/worker队列
		maxWorker := 100
		maxQueue := 100
		d := NewDispatcher(maxWorker)
		d.Run(maxWorker, maxQueue)
	})
}