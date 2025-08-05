package actor

import "time"

const (
	WAIT_GAME_FINISH = 1 << 0 // 等待游戏结束
	WAIT_FINISH      = 1 << 1 // 等待结束
	FINISHED         = 1 << 2 // 已经结束
)

type BaseActor struct {
	Actor
	update int64  // 更新时间
	change uint32 // 变更次数
	flag   uint32 // 结束状态
}

// 销毁actor状态
func (d *BaseActor) WaitGameFinish() {
	d.flag = WAIT_GAME_FINISH
}

func (d *BaseActor) WaitFinish() {
	d.flag = WAIT_FINISH
}

func (d *BaseActor) Finished() {
	d.flag = FINISHED
}

// 是否等待游戏结束
func (d *BaseActor) IsWaitGameFinish() bool {
	return d.flag == WAIT_GAME_FINISH
}

// 是否等待结束
func (d *BaseActor) IsWaitFinish() bool {
	return d.flag == WAIT_FINISH
}

// 是否已经结束
func (d *BaseActor) IsFinished() bool {
	return d.flag == FINISHED
}

func (d *BaseActor) GetUpdateTime() int64 {
	return d.update
}

func (d *BaseActor) SetUpdateTime(now int64) {
	d.update = now
}

// 保存
func (d *BaseActor) Flush() {
	d.change = 0
	d.SetUpdateTime(time.Now().Unix())
}

// 是否变更
func (d *BaseActor) IsChange() bool {
	return d.change > 0
}

// 变更
func (d *BaseActor) Change() {
	d.change++
}
