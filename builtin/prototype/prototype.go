package prototype

import (
	"math/big"

	"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
)

type Prototype struct {
	addr  thor.Address
	state *state.State
}

func New(addr thor.Address, state *state.State) *Prototype {
	return &Prototype{addr, state}
}

func (p *Prototype) Bind(target thor.Address) *Binding {
	return &Binding{p.addr, p.state, target}
}

type Binding struct {
	selfAddr thor.Address
	state    *state.State
	target   thor.Address
}

func (b *Binding) masterKey() thor.Bytes32 {
	return thor.Blake2b(b.target.Bytes(), []byte("master"))
}

func (b *Binding) userKey(user thor.Address) thor.Bytes32 {
	return thor.Blake2b(b.target.Bytes(), user.Bytes(), []byte("user"))
}
func (b *Binding) userPlanKey() thor.Bytes32 {
	return thor.Blake2b(b.target.Bytes(), []byte("user-plan"))
}

func (b *Binding) sponsorKey(sponsor thor.Address) thor.Bytes32 {
	return thor.Blake2b(b.target.Bytes(), sponsor.Bytes(), []byte("sponsor"))
}

func (b *Binding) curSponsorKey() thor.Bytes32 {
	return thor.Blake2b(b.target.Bytes(), []byte("cur-sponsor"))
}

func (b *Binding) getStorage(key thor.Bytes32, val interface{}) {
	b.state.GetStructedStorage(b.selfAddr, key, val)
}

func (b *Binding) setStorage(key thor.Bytes32, val interface{}) {
	b.state.SetStructedStorage(b.selfAddr, key, val)
}

func (b *Binding) Master() (master thor.Address) {
	b.getStorage(b.masterKey(), &master)
	return
}

func (b *Binding) SetMaster(master thor.Address) {
	b.setStorage(b.masterKey(), &master)
}

func (b *Binding) IsUser(user thor.Address) bool {
	var uo userObject
	b.getStorage(b.userKey(user), &uo)
	return !uo.IsEmpty()
}

func (b *Binding) AddUser(user thor.Address, blockNum uint32) bool {
	userKey := b.userKey(user)
	var uo userObject
	b.getStorage(userKey, &uo)

	if !uo.IsEmpty() {
		return false
	}

	var up userPlan
	b.getStorage(b.userPlanKey(), &up)

	b.setStorage(userKey, &userObject{
		up.Credit,
		blockNum,
	})
	return true
}

func (b *Binding) RemoveUser(user thor.Address) bool {
	userKey := b.userKey(user)
	var uo userObject
	b.getStorage(userKey, &uo)
	if uo.IsEmpty() {
		return false
	}
	// clear storage
	b.setStorage(userKey, uint8(0))
	return true
}

func (b *Binding) UserCredit(user thor.Address, blockNum uint32) *big.Int {
	var uo userObject
	b.getStorage(b.userKey(user), &uo)
	if uo.IsEmpty() {
		return &big.Int{}
	}
	var up userPlan
	b.getStorage(b.userPlanKey(), &up)
	return uo.Credit(&up, blockNum)
}

func (b *Binding) SetUserCredit(user thor.Address, credit *big.Int, blockNum uint32) {
	b.setStorage(b.userKey(user), &userObject{credit, blockNum})
}

func (b *Binding) UserPlan() (credit, recoveryRate *big.Int) {
	var up userPlan
	b.getStorage(b.userPlanKey(), &up)
	return up.Credit, up.RecoveryRate
}

func (b *Binding) SetUserPlan(credit, recoveryRate *big.Int) {
	b.setStorage(b.userPlanKey(), &userPlan{credit, recoveryRate})
}

func (b *Binding) Sponsor(sponsor thor.Address, yesOrNo bool) bool {
	sponsorKey := b.sponsorKey(sponsor)
	var flag uint8
	b.getStorage(sponsorKey, &flag)
	if yesOrNo {
		if flag != 0 {
			return false
		}
		b.setStorage(sponsorKey, uint8(1))
	} else {
		if flag == 0 {
			return false
		}
		b.setStorage(sponsorKey, uint8(0))
		if b.CurrentSponsor() == sponsor {
			b.setStorage(b.curSponsorKey(), thor.Address{})
		}
	}
	return true
}

func (b *Binding) IsSponsor(sponsor thor.Address) bool {
	var flag uint8
	b.getStorage(b.sponsorKey(sponsor), &flag)
	return flag != 0
}

func (b *Binding) SelectSponsor(sponsor thor.Address) bool {
	if sponsor.IsZero() {
		// allow select zero sponsor to clear current
		b.setStorage(b.curSponsorKey(), sponsor)
		return true
	}

	if !b.IsSponsor(sponsor) {
		return false
	}
	b.setStorage(b.curSponsorKey(), sponsor)
	return true
}

func (b *Binding) CurrentSponsor() (addr thor.Address) {
	b.getStorage(b.curSponsorKey(), &addr)
	return
}