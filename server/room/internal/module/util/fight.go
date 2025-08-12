package util

import (
	"go-actor/common/pb"
	"go-actor/library/random"
)

// GetSkillById 提取技能
func GetSkillById(dst uint32, src []*pb.Skill) *pb.Skill {
	for i := range src {
		if src[i].ID == dst {
			return src[i]
		}
	}
	return nil
}

func GetAliveEnemy(enemys []*pb.Enemys) (alive []*pb.Enemys) {
	alive = make([]*pb.Enemys, 0, len(enemys))
	for i := range enemys {
		if enemys[i].Status == pb.Status_StatusAlive {
			alive = append(alive, enemys[i])
		}
	}
	return
}

func RandAliveEnemy(enemys []*pb.Enemys, length uint32) (ret []*pb.Enemys) {
	alives := GetAliveEnemy(enemys)
	if int(length) >= len(alives) {
		return alives
	}
	rands := random.Perm(len(alives))
	for i := range rands {
		ret = append(ret, enemys[rands[i]])
	}
	return
}

func GetAliveCharacters(characters []*pb.Character) (alive []*pb.Character) {
	alive = make([]*pb.Character, 0, len(characters))
	for i := range characters {
		if characters[i].Status == pb.Status_StatusAlive {
			alive = append(alive, characters[i])
		}
	}
	return
}

func RandAliveCharacters(characters []*pb.Character, length uint32) (ret []*pb.Character) {
	alives := GetAliveCharacters(characters)
	if int(length) >= len(alives) {
		return alives
	}
	rands := random.Perm(len(alives))
	for i := range rands {
		ret = append(ret, characters[rands[i]])
	}
	return
}

// Active 激发技能 todo
func Active(event pb.TriggerType, src []*pb.Skill) {

}
