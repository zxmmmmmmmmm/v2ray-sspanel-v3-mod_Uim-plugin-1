package db

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/zxmmmmmmmmm/v2ray-sspanel-v3-mod_Uim-plugin-1/model"
	"github.com/zxmmmmmmmmm/v2ray-sspanel-v3-mod_Uim-plugin-1/speedtest"
	"github.com/zxmmmmmmmmm/v2ray-sspanel-v3-mod_Uim-plugin-1/utility"
	"log"
	"strconv"
	"strings"
	"time"
)

type Webapi struct {
	WebToken   string
	WebBaseURl string
	MU_REGEX   string
	MU_SUFFIX  string
}

func (api *Webapi) GetApi(url string, params map[string]interface{}) (*req.Resp, error) {
	req.SetTimeout(50 * time.Second)
	parm := req.Param{
		"key": api.WebToken,
	}
	for k, v := range params {
		parm[k] = v
	}

	r, err := req.Get(fmt.Sprintf("%s/mod_mu/%s", api.WebBaseURl, url), parm)
	return r, err
}

func (api *Webapi) GetNodeInfo(nodeid uint) (*NodeinfoResponse, error) {
	var response = NodeinfoResponse{}
	var params map[string]interface{}

	r, err := api.GetApi(fmt.Sprintf("nodes/%d/info", nodeid), params)
	if err != nil {
		return &response, err
	} else {
		err = r.ToJSON(&response)
		if err != nil {
			return &response, err
		} else if response.Ret != 1 {
			return &response, err
		}
	}

	if response.Data.Server_raw != "" {
		response.Data.Server_raw = strings.ToLower(response.Data.Server_raw)
		data := strings.Split(response.Data.Server_raw, ";")
		var count uint
		count = 0
		for v := range data {
			if len(data[v]) > 0 {
				maps[id2string[count]] = data[v]
			}
			count += 1
		}
		var extraArgues []string
		if len(data) == 6 {
			extraArgues = append(extraArgues, strings.Split(data[5], "|")...)
			for item := range extraArgues {
				data = strings.Split(extraArgues[item], "=")
				if len(data) > 0 {
					if len(data[1]) > 0 {
						maps[data[0]] = data[1]
					}

				}
			}
		}

		if maps["protocol"] == "tls" {
			temp := maps["protocol_param"]
			maps["protocol"] = temp
			maps["protocol_param"] = "tls"
		}
		response.Data.Server = maps
	}
	response.Data.NodeID = nodeid
	return &response, nil
}

func (api *Webapi) GetDisNodeInfo(nodeid uint) (*DisNodenfoResponse, error) {
	var response = DisNodenfoResponse{}
	var params map[string]interface{}
	params = map[string]interface{}{
		"node_id": nodeid,
	}
	r, err := api.GetApi("func/relay_rules", params)
	if err != nil {
		return &response, err
	} else {
		err = r.ToJSON(&response)
		if err != nil {
			return &response, err
		} else if response.Ret != 1 {
			return &response, err
		}
	}

	if len(response.Data) > 0 {
		for _, relayrule := range response.Data {
			relayrule.Server_raw = strings.ToLower(relayrule.Server_raw)
			data := strings.Split(relayrule.Server_raw, ";")
			var count uint
			count = 0
			for v := range data {
				if len(data[v]) > 1 {
					maps[id2string[count]] = data[v]
				}
				count += 1
			}
			var extraArgues []string
			if len(data) == 6 {
				extraArgues = append(extraArgues, strings.Split(data[5], "|")...)
				for item := range extraArgues {
					data = strings.Split(extraArgues[item], "=")
					if len(data) > 1 {
						if len(data[1]) > 1 {
							maps[data[0]] = data[1]
						}

					}
				}
			}

			if maps["protocol"] == "tls" {
				temp := maps["protocol_param"]
				maps["protocol"] = temp
				maps["protocol_param"] = "tls"
			}
			relayrule.Server = maps
		}
	}
	return &response, nil
}

func (api *Webapi) GetALLUsers(info *model.NodeInfo) (*AllUsers, error) {
	sort := info.Sort
	var prifix string
	var allusers = AllUsers{
		Data: map[string]model.UserModel{},
	}
	if sort == 0 {
		prifix = "SS_"
	} else {
		prifix = "Vmess_"
		if info.Server["protocol"] == "tcp" {
			prifix += "tcp_"
		} else if info.Server["protocol"] == "ws" {
			if info.Server["protocol_param"] != "" {
				prifix += "ws_" + info.Server["protocol_param"].(string) + "_"
			} else {
				prifix += "ws_" + "none" + "_"
			}
		} else if info.Server["protocol"] == "kcp" {
			if info.Server["protocol_param"] != "" {
				prifix += "kcp_" + info.Server["protocol_param"].(string) + "_"
			} else {
				prifix += "kcp_" + "none" + "_"
			}
		}
	}
	var response = UsersResponse{}
	params := map[string]interface{}{
		"node_id": info.NodeID,
	}
	r, err := api.GetApi("users", params)
	if err != nil {
		return &allusers, err
	} else {
		err = r.ToJSON(&response)
		allusers.Ret = response.Ret
		if err != nil {
			return &allusers, err
		} else if response.Ret != 1 {
			return &allusers, err
		}
	}
	for index := range response.Data {
		// 按照node 限速来调整用户限速
		if info.NodeSpeedlimit != 0 {
			if info.NodeSpeedlimit > response.Data[index].NodeSpeedlimit && response.Data[index].NodeSpeedlimit != 0 {
			} else if response.Data[index].NodeSpeedlimit == 0 || response.Data[index].NodeSpeedlimit > info.NodeSpeedlimit {
				response.Data[index].NodeSpeedlimit = info.NodeSpeedlimit
			}
		}
		// 接受到的是 Mbps， 然后我们的一个buffer 是2048byte， 差不多61个
		response.Data[index].Rate = uint32(response.Data[index].NodeSpeedlimit * 62)
		if info.Server["alterid"].(string) == "" {
			response.Data[index].AlterId = 2
		} else {
			alterid, err := strconv.ParseUint(info.Server["alterid"].(string), 10, 0)
			if err == nil {
				response.Data[index].AlterId = uint32(alterid)
			}
		}
		user := response.Data[index]
		response.Data[index].Muhost = get_mu_host(user.UserID, getMD5(fmt.Sprintf("%d%s%s%s%s", user.UserID, user.Passwd, user.Method, user.Obfs, user.Protocol)), api.MU_REGEX, api.MU_SUFFIX)
		key := prifix + response.Data[index].Email + fmt.Sprintf("Rate_%d_AlterID_%d_Method_%s_Passwd_%s_Port_%d_Obfs_%s_Protocol_%s", response.Data[index].Rate,
			response.Data[index].AlterId, response.Data[index].Method, response.Data[index].Passwd, response.Data[index].Port, response.Data[index].Obfs, response.Data[index].Protocol,
		)
		response.Data[index].PrefixedId = key
		allusers.Data[key] = response.Data[index]
	}
	return &allusers, nil
}

func (api *Webapi) Post(url string, params map[string]interface{}, data map[string]interface{}) (*req.Resp, error) {
	parm := req.Param{
		"key": api.WebToken,
	}
	for k, v := range params {
		parm[k] = v
	}
	r, err := req.Post(fmt.Sprintf("%s/mod_mu/%s", api.WebBaseURl, url), parm, req.BodyJSON(&data))
	return r, err
}

func (api *Webapi) UploadSystemLoad(nodeid uint) bool {
	var postresponse PostResponse
	params := map[string]interface{}{
		"node_id": nodeid,
	}
	upload_systemload := map[string]interface{}{
		"uptime": utility.GetSystemUptime(),
		"load":   utility.GetSystemLoad(),
	}
	r, err := api.Post(fmt.Sprintf("nodes/%d/info", nodeid), params, upload_systemload)
	if err != nil {
		return false
	} else {
		err = r.ToJSON(&postresponse)
		if err != nil {
			return false
		} else if postresponse.Ret != 1 {
			log.Fatal(postresponse.Data)
		}
	}
	return true
}

func (api *Webapi) UpLoadUserTraffics(nodeid uint, trafficLog []model.UserTrafficLog) bool {
	var postresponse PostResponse
	params := map[string]interface{}{
		"node_id": nodeid,
	}

	data := map[string]interface{}{
		"data": trafficLog,
	}
	r, err := api.Post("users/traffic", params, data)
	if err != nil {
		return false
	} else {
		err = r.ToJSON(&postresponse)
		if err != nil {
			return false
		} else if postresponse.Ret != 1 {
			log.Fatal(postresponse.Data)
		}
	}
	return true
}
func (api *Webapi) UploadSpeedTest(nodeid uint, speedresult []speedtest.Speedresult) bool {
	var postresponse PostResponse
	params := map[string]interface{}{
		"node_id": nodeid,
	}

	data := map[string]interface{}{
		"data": speedresult,
	}
	r, err := api.Post("func/speedtest", params, data)
	if err != nil {
		return false
	} else {
		err = r.ToJSON(&postresponse)
		if err != nil {
			return false
		} else if postresponse.Ret != 1 {
			log.Fatal(postresponse.Data)
		}
	}
	return true
}
func (api *Webapi) UpLoadOnlineIps(nodeid uint, onlineIPS []model.UserOnLineIP) bool {
	var postresponse PostResponse
	params := map[string]interface{}{
		"node_id": nodeid,
	}

	data := map[string]interface{}{
		"data": onlineIPS,
	}
	r, err := api.Post("users/aliveip", params, data)
	if err != nil {
		return false
	} else {
		err = r.ToJSON(&postresponse)
		if err != nil {
			return false
		} else if postresponse.Ret != 1 {
			log.Fatal(postresponse.Data)
		}
	}
	return true
}
