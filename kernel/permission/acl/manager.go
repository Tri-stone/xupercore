package acl

import (
	"encoding/json"
	"errors"

	"github.com/xuperchain/xupercore/kernel/permission/acl/base"
	actx "github.com/xuperchain/xupercore/kernel/permission/acl/context"
	"github.com/xuperchain/xupercore/kernel/permission/acl/utils"
	pb "github.com/xuperchain/xupercore/protos"
)

// Manager manages all ACL releated data, providing read/write interface for ACL table
type Manager struct {
	Ctx *actx.AclCtx
}

// NewACLManager create instance of ACLManager
func NewACLManager(ctx *actx.AclCtx) (base.AclManager, error) {
	if ctx == nil || ctx.Ledger == nil || ctx.Register == nil || ctx.BcName == "" {
		return nil, fmt.Errorf("acl ctx set error")
	}

	newAccountGas, err := ctx.Ledger.GetNewAccountGas()
	if err != nil {
		return nil, fmt.Errorf("get create account gas failed.err:%v", err)
	}

	t := NewKernContractMethod(ctx.BcName, newAccountGas)
	ctx.Register.RegisterKernMethod(utils.SubModName, "NewAccount", t.NewAccount)
	ctx.Register.RegisterKernMethod(utils.SubModName, "SetAccountACL", t.SetAccountACL)
	ctx.Register.RegisterKernMethod(utils.SubModName, "SetMethodACL", t.SetMethodACL)

	mg := &Manager{
		Ctx: ctx,
	}

	return mg, nil
}

// GetAccountACL get acl of an account
func (mgr *Manager) GetAccountACL(accountName string) (*pb.Acl, error) {
	acl, err := t.GetObjectBySnapshot(utils.GetAccountBucket(), accountName)
	if err != nil {
		return nil, fmt.Errorf("query account acl failed.err:%v", err)
	}

	aclBuf := &pb.Acl{}
	err = json.Unmarshal(acl, aclBuf)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal acl failed.acl:%s,err:%v", string(acl), err)
	}
	return aclBuf, nil
}

// GetContractMethodACL get acl of contract method
func (mgr *Manager) GetContractMethodACL(contractName, methodName string) (*pb.Acl, error) {
	key := utils.MakeContractMethodKey(contractName, methodName)
	acl, err := t.GetObjectBySnapshot(utils.GetContractBucket(), key)
	if err != nil {
		return nil, fmt.Errorf("query contract method acl failed.err:%v", err)
	}

	aclBuf := &pb.Acl{}
	err = json.Unmarshal(acl, aclBuf)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal acl failed.acl:%s,err:%v", string(acl), err)
	}
	return aclBuf, nil
}

// GetAccountAddresses get the addresses belongs to contract account
func (mgr *Manager) GetAccountAddresses(accountName string) ([]string, error) {
	acl, err := mgr.GetAccountACL(accountName)
	if err != nil {
		return nil, err
	}

	return mgr.getAddressesByACL(acl)
}

func (mgr *Manager) GetObjectBySnapshot(bucket string, object []byte) ([]byte, error) {
	// 根据tip blockid 创建快照
	reader, err := mgr.Ctx.Ledger.GetTipXMSnapshotReader()
	if err != nil {
		return nil, err
	}

	return reader.Get(bucket, object)
}

func (mgr *Manager) getAddressesByACL(acl *pb.Acl) ([]string, error) {
	addresses := make([]string, 0)

	switch acl.GetPm().GetRule() {
	case pb.PermissionRule_SIGN_THRESHOLD:
		for ak := range acl.GetAksWeight() {
			addresses = append(addresses, ak)
		}
	case pb.PermissionRule_SIGN_AKSET:
		for _, set := range acl.GetAkSets().GetSets() {
			aks := set.GetAks()
			addresses = append(addresses, aks...)
		}
	default:
		return nil, errors.New("Unknown permission rule")
	}

	return addresses, nil
}