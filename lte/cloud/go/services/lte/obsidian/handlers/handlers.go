/*
 * Copyright 2020 The Magma Authors.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handlers

import (
	"fmt"
	"net/http"

	"magma/lte/cloud/go/lte"
	lte_models "magma/lte/cloud/go/services/lte/obsidian/models"
	policydb_models "magma/lte/cloud/go/services/policydb/obsidian/models"
	"magma/orc8r/cloud/go/models"
	"magma/orc8r/cloud/go/obsidian"
	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/services/configurator"
	"magma/orc8r/cloud/go/services/orchestrator/obsidian/handlers"
	orc8r_models "magma/orc8r/cloud/go/services/orchestrator/obsidian/models"
	"magma/orc8r/cloud/go/services/state"
	"magma/orc8r/cloud/go/storage"
	merrors "magma/orc8r/lib/go/errors"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

const (
	LteNetworks                         = "lte"
	ListNetworksPath                    = obsidian.V1Root + LteNetworks
	ManageNetworkPath                   = ListNetworksPath + obsidian.UrlSep + ":network_id"
	ManageNetworkNamePath               = ManageNetworkPath + obsidian.UrlSep + "name"
	ManageNetworkDescriptionPath        = ManageNetworkPath + obsidian.UrlSep + "description"
	ManageNetworkFeaturesPath           = ManageNetworkPath + obsidian.UrlSep + "features"
	ManageNetworkDNSPath                = ManageNetworkPath + obsidian.UrlSep + "dns"
	ManageNetworkDNSRecordsPath         = ManageNetworkDNSPath + obsidian.UrlSep + "records"
	ManageNetworkDNSRecordByDomainPath  = ManageNetworkDNSRecordsPath + obsidian.UrlSep + ":domain"
	ManageNetworkCellularPath           = ManageNetworkPath + obsidian.UrlSep + "cellular"
	ManageNetworkCellularEpcPath        = ManageNetworkCellularPath + obsidian.UrlSep + "epc"
	ManageNetworkCellularRanPath        = ManageNetworkCellularPath + obsidian.UrlSep + "ran"
	ManageNetworkCellularFegNetworkID   = ManageNetworkCellularPath + obsidian.UrlSep + "feg_network_id"
	ManageNetworkSubscriberPath         = ManageNetworkPath + obsidian.UrlSep + "subscriber_config"
	ManageNetworkBaseNamesPath          = ManageNetworkSubscriberPath + obsidian.UrlSep + "base_names"
	ManageNetworkRuleNamesPath          = ManageNetworkSubscriberPath + obsidian.UrlSep + "rule_names"
	ManageNetworkSubscriberRuleNamePath = ManageNetworkRuleNamesPath + obsidian.UrlSep + ":rule_id"
	ManageNetworkSubscriberBaseNamePath = ManageNetworkBaseNamesPath + obsidian.UrlSep + ":base_name"

	ManageNetworkApnPath              = ManageNetworkPath + obsidian.UrlSep + "apns"
	ManageNetworkApnConfigurationPath = ManageNetworkApnPath + obsidian.UrlSep + ":apn_name"

	Gateways                          = "gateways"
	ListGatewaysPath                  = ManageNetworkPath + obsidian.UrlSep + Gateways
	ManageGatewayPath                 = ListGatewaysPath + obsidian.UrlSep + ":gateway_id"
	ManageGatewayNamePath             = ManageGatewayPath + obsidian.UrlSep + "name"
	ManageGatewayDescriptionPath      = ManageGatewayPath + obsidian.UrlSep + "description"
	ManageGatewayConfigPath           = ManageGatewayPath + obsidian.UrlSep + "magmad"
	ManageGatewayDevicePath           = ManageGatewayPath + obsidian.UrlSep + "device"
	ManageGatewayStatePath            = ManageGatewayPath + obsidian.UrlSep + "status"
	ManageGatewayTierPath             = ManageGatewayPath + obsidian.UrlSep + "tier"
	ManageGatewayCellularPath         = ManageGatewayPath + obsidian.UrlSep + "cellular"
	ManageGatewayCellularEpcPath      = ManageGatewayCellularPath + obsidian.UrlSep + "epc"
	ManageGatewayCellularRanPath      = ManageGatewayCellularPath + obsidian.UrlSep + "ran"
	ManageGatewayCellularNonEpsPath   = ManageGatewayCellularPath + obsidian.UrlSep + "non_eps"
	ManageGatewayCellularDNSPath      = ManageGatewayCellularPath + obsidian.UrlSep + "dns"
	ManageGatewayDNSRecordsPath       = ManageGatewayCellularDNSPath + obsidian.UrlSep + "records"
	ManageGatewayConnectedEnodebsPath = ManageGatewayPath + obsidian.UrlSep + "connected_enodeb_serials"
	ManageGatewayVPNConfigPath        = ManageGatewayPath + obsidian.UrlSep + "vpn"

	Enodebs            = "enodebs"
	ListEnodebsPath    = ManageNetworkPath + obsidian.UrlSep + Enodebs
	ManageEnodebPath   = ListEnodebsPath + obsidian.UrlSep + ":enodeb_serial"
	GetEnodebStatePath = ManageEnodebPath + obsidian.UrlSep + "state"
)

func GetHandlers() []obsidian.Handler {
	ret := []obsidian.Handler{
		{Path: ManageNetworkDNSRecordByDomainPath, Methods: obsidian.POST, HandlerFunc: handlers.CreateDNSRecord},
		{Path: ManageNetworkDNSRecordByDomainPath, Methods: obsidian.GET, HandlerFunc: handlers.ReadDNSRecord},
		{Path: ManageNetworkDNSRecordByDomainPath, Methods: obsidian.PUT, HandlerFunc: handlers.UpdateDNSRecord},
		{Path: ManageNetworkDNSRecordByDomainPath, Methods: obsidian.DELETE, HandlerFunc: handlers.DeleteDNSRecord},

		handlers.GetListGatewaysHandler(ListGatewaysPath, &lte_models.MutableLteGateway{}, makeLTEGateways),
		{Path: ListGatewaysPath, Methods: obsidian.POST, HandlerFunc: createGateway},
		{Path: ManageGatewayPath, Methods: obsidian.GET, HandlerFunc: getGateway},
		{Path: ManageGatewayPath, Methods: obsidian.PUT, HandlerFunc: updateGateway},
		{Path: ManageGatewayPath, Methods: obsidian.DELETE, HandlerFunc: deleteGateway},

		{Path: ManageGatewayStatePath, Methods: obsidian.GET, HandlerFunc: handlers.GetStateHandler},

		{Path: ListEnodebsPath, Methods: obsidian.GET, HandlerFunc: listEnodebs},
		{Path: ListEnodebsPath, Methods: obsidian.POST, HandlerFunc: createEnodeb},
		{Path: ManageEnodebPath, Methods: obsidian.GET, HandlerFunc: getEnodeb},
		{Path: ManageEnodebPath, Methods: obsidian.PUT, HandlerFunc: updateEnodeb},
		{Path: ManageEnodebPath, Methods: obsidian.DELETE, HandlerFunc: deleteEnodeb},
		{Path: ManageGatewayConnectedEnodebsPath, Methods: obsidian.POST, HandlerFunc: addConnectedEnodeb},
		{Path: ManageGatewayConnectedEnodebsPath, Methods: obsidian.DELETE, HandlerFunc: deleteConnectedEnodeb},
		{Path: GetEnodebStatePath, Methods: obsidian.GET, HandlerFunc: getEnodebState},

		{Path: ManageNetworkApnPath, Methods: obsidian.GET, HandlerFunc: listApns},
		{Path: ManageNetworkApnPath, Methods: obsidian.POST, HandlerFunc: createApn},
		{Path: ManageNetworkApnConfigurationPath, Methods: obsidian.GET, HandlerFunc: getApnConfiguration},
		{Path: ManageNetworkApnConfigurationPath, Methods: obsidian.PUT, HandlerFunc: updateApnConfiguration},
		{Path: ManageNetworkApnConfigurationPath, Methods: obsidian.DELETE, HandlerFunc: deleteApnConfiguration},

		{Path: ManageNetworkSubscriberBaseNamePath, Methods: obsidian.POST, HandlerFunc: AddNetworkWideSubscriberBaseName},
		{Path: ManageNetworkSubscriberRuleNamePath, Methods: obsidian.POST, HandlerFunc: AddNetworkWideSubscriberRuleName},
		{Path: ManageNetworkSubscriberBaseNamePath, Methods: obsidian.DELETE, HandlerFunc: RemoveNetworkWideSubscriberBaseName},
		{Path: ManageNetworkSubscriberRuleNamePath, Methods: obsidian.DELETE, HandlerFunc: RemoveNetworkWideSubscriberRuleName},
	}
	ret = append(ret, handlers.GetTypedNetworkCRUDHandlers(ListNetworksPath, ManageNetworkPath, lte.NetworkType, &lte_models.LteNetwork{})...)

	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkNamePath, new(models.NetworkName), "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkDescriptionPath, new(models.NetworkDescription), "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkFeaturesPath, &orc8r_models.NetworkFeatures{}, orc8r.NetworkFeaturesConfig)...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkDNSPath, &orc8r_models.NetworkDNSConfig{}, orc8r.DnsdNetworkType)...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkDNSRecordsPath, new(orc8r_models.NetworkDNSRecords), "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkCellularPath, &lte_models.NetworkCellularConfigs{}, lte.CellularNetworkConfigType)...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkCellularEpcPath, &lte_models.NetworkEpcConfigs{}, "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkCellularRanPath, &lte_models.NetworkRanConfigs{}, "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkCellularFegNetworkID, new(lte_models.FegNetworkID), "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkSubscriberPath, &policydb_models.NetworkSubscriberConfig{}, "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkRuleNamesPath, new(policydb_models.RuleNames), "")...)
	ret = append(ret, handlers.GetPartialNetworkHandlers(ManageNetworkBaseNamesPath, new(policydb_models.BaseNames), "")...)

	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayNamePath, new(models.GatewayName))...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayDescriptionPath, new(models.GatewayDescription))...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayConfigPath, &orc8r_models.MagmadGatewayConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayTierPath, new(orc8r_models.TierID))...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayCellularPath, &lte_models.GatewayCellularConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayCellularEpcPath, &lte_models.GatewayEpcConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayCellularRanPath, &lte_models.GatewayRanConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayCellularNonEpsPath, &lte_models.GatewayNonEpsConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayCellularDNSPath, &lte_models.GatewayDNSConfigs{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayDNSRecordsPath, &lte_models.GatewayDNSRecords{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayConnectedEnodebsPath, &lte_models.EnodebSerials{})...)
	ret = append(ret, handlers.GetPartialGatewayHandlers(ManageGatewayVPNConfigPath, &orc8r_models.GatewayVpnConfigs{})...)

	ret = append(ret, handlers.GetGatewayDeviceHandlers(ManageGatewayDevicePath)...)

	return ret
}

func createGateway(c echo.Context) error {
	if nerr := handlers.CreateGateway(c, &lte_models.MutableLteGateway{}); nerr != nil {
		return nerr
	}
	return c.NoContent(http.StatusCreated)
}

func getGateway(c echo.Context) error {
	nid, gid, nerr := obsidian.GetNetworkAndGatewayIDs(c)
	if nerr != nil {
		return nerr
	}

	magmadModel, nerr := handlers.LoadMagmadGateway(nid, gid)
	if nerr != nil {
		return nerr
	}

	ent, err := configurator.LoadEntity(
		nid, lte.CellularGatewayEntityType, gid,
		configurator.EntityLoadCriteria{LoadConfig: true, LoadAssocsFromThis: true},
	)
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "failed to load cellular gateway"), http.StatusInternalServerError)
	}

	ret := &lte_models.LteGateway{
		ID:                     magmadModel.ID,
		Name:                   magmadModel.Name,
		Description:            magmadModel.Description,
		Device:                 magmadModel.Device,
		Status:                 magmadModel.Status,
		Tier:                   magmadModel.Tier,
		Magmad:                 magmadModel.Magmad,
		ConnectedEnodebSerials: lte_models.EnodebSerials{},
		ApnResources:           lte_models.ApnResources{},
	}
	if ent.Config != nil {
		ret.Cellular = ent.Config.(*lte_models.GatewayCellularConfigs)
	}

	for _, tk := range ent.Associations {
		switch tk.Type {
		case lte.CellularEnodebEntityType:
			ret.ConnectedEnodebSerials = append(ret.ConnectedEnodebSerials, tk.Key)
		case lte.APNResourceEntityType:
			e, err := configurator.LoadEntity(nid, tk.Type, tk.Key, configurator.EntityLoadCriteria{LoadConfig: true})
			if err != nil {
				return errors.Wrap(err, "error loading apn resource entity")
			}
			apnResource := (&lte_models.ApnResource{}).FromEntity(e)
			ret.ApnResources[string(apnResource.ApnName)] = *apnResource
		}
	}

	return c.JSON(http.StatusOK, ret)
}

func updateGateway(c echo.Context) error {
	nid, gid, nerr := obsidian.GetNetworkAndGatewayIDs(c)
	if nerr != nil {
		return nerr
	}
	if nerr = handlers.UpdateGateway(c, nid, gid, &lte_models.MutableLteGateway{}); nerr != nil {
		return nerr
	}
	return c.NoContent(http.StatusNoContent)
}

func deleteGateway(c echo.Context) error {
	nid, gid, nerr := obsidian.GetNetworkAndGatewayIDs(c)
	if nerr != nil {
		return nerr
	}

	var deletes storage.TKs
	deletes = append(deletes, storage.TypeAndKey{Type: lte.CellularGatewayEntityType, Key: gid})

	gw, err := configurator.LoadEntity(
		nid, lte.CellularGatewayEntityType, gid,
		configurator.EntityLoadCriteria{LoadAssocsFromThis: true},
	)
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "error loading existing cellular gateway"), http.StatusInternalServerError)
	}
	deletes = append(deletes, gw.Associations.Filter(lte.APNResourceEntityType)...)

	err = handlers.DeleteMagmadGateway(nid, gid, deletes)
	if err != nil {
		return makeErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

type cellularAndMagmadGateway struct {
	magmadGateway, cellularGateway configurator.NetworkEntity
}

func makeLTEGateways(
	entsByTK configurator.NetworkEntitiesByTK,
	devicesByID map[string]interface{},
	statusesByID map[string]*orc8r_models.GatewayStatus,
) map[string]handlers.GatewayModel {
	gatewayEntsByKey := map[string]*cellularAndMagmadGateway{}
	for tk, ent := range entsByTK.MultiFilter(orc8r.MagmadGatewayType, lte.CellularGatewayEntityType) {
		existing, found := gatewayEntsByKey[tk.Key]
		if !found {
			existing = &cellularAndMagmadGateway{}
			gatewayEntsByKey[tk.Key] = existing
		}
		switch ent.Type {
		case orc8r.MagmadGatewayType:
			existing.magmadGateway = ent
		case lte.CellularGatewayEntityType:
			existing.cellularGateway = ent
		}
	}

	cellularGateways := map[string]handlers.GatewayModel{}
	for key, ents := range gatewayEntsByKey {
		hwID := ents.magmadGateway.PhysicalID
		var devCasted *orc8r_models.GatewayDevice
		if devicesByID[hwID] != nil {
			devCasted = devicesByID[hwID].(*orc8r_models.GatewayDevice)
		}
		cellularGateways[key] = (&lte_models.LteGateway{}).FromBackendModels(ents.magmadGateway, ents.cellularGateway, entsByTK, devCasted, statusesByID[hwID])
	}
	return cellularGateways
}

func listEnodebs(c echo.Context) error {
	nid, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}

	ents, err := configurator.LoadAllEntitiesInNetwork(
		nid, lte.CellularEnodebEntityType,
		configurator.EntityLoadCriteria{LoadMetadata: true, LoadConfig: true, LoadAssocsToThis: true},
	)
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	ret := make(map[string]*lte_models.Enodeb, len(ents))
	for _, ent := range ents {
		ret[ent.Key] = (&lte_models.Enodeb{}).FromBackendModels(ent)
	}
	return c.JSON(http.StatusOK, ret)
}

func createEnodeb(c echo.Context) error {
	nid, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}

	payload := &lte_models.Enodeb{}
	if err := c.Bind(payload); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if err := payload.ValidateModel(); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if payload.AttachedGatewayID != "" {
		return echo.NewHTTPError(http.StatusBadRequest, "attached_gateway_id is a read-only property")
	}

	_, err := configurator.CreateEntity(nid, configurator.NetworkEntity{
		Type:        lte.CellularEnodebEntityType,
		Key:         payload.Serial,
		Name:        payload.Name,
		Description: payload.Description,
		PhysicalID:  payload.Serial,
		Config:      payload.Config,
	})
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func getEnodeb(c echo.Context) error {
	nid, eid, nerr := getNetworkAndEnbIDs(c)
	if nerr != nil {
		return nerr
	}

	ent, err := configurator.LoadEntity(
		nid, lte.CellularEnodebEntityType, eid,
		configurator.EntityLoadCriteria{LoadMetadata: true, LoadConfig: true, LoadAssocsToThis: true},
	)
	switch {
	case err == merrors.ErrNotFound:
		return echo.ErrNotFound
	case err != nil:
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	ret := (&lte_models.Enodeb{}).FromBackendModels(ent)
	return c.JSON(http.StatusOK, ret)
}

func updateEnodeb(c echo.Context) error {
	nid, eid, nerr := getNetworkAndEnbIDs(c)
	if nerr != nil {
		return nerr
	}

	payload := &lte_models.Enodeb{}
	if err := c.Bind(payload); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if err := payload.ValidateModel(); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if payload.AttachedGatewayID != "" {
		return echo.NewHTTPError(http.StatusBadRequest, "attached_gateway_id is a read-only property")
	}
	if payload.Serial != eid {
		return echo.NewHTTPError(http.StatusBadRequest, "serial in body must match serial in path")
	}

	_, err := configurator.UpdateEntity(nid, payload.ToEntityUpdateCriteria())
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func deleteEnodeb(c echo.Context) error {
	nid, eid, nerr := getNetworkAndEnbIDs(c)
	if nerr != nil {
		return nerr
	}

	err := configurator.DeleteEntity(nid, lte.CellularEnodebEntityType, eid)
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func getEnodebState(c echo.Context) error {
	nid, eid, nerr := getNetworkAndEnbIDs(c)
	if nerr != nil {
		return nerr
	}
	st, err := state.GetState(nid, lte.EnodebStateType, eid)
	if err == merrors.ErrNotFound {
		return obsidian.HttpError(err, http.StatusNotFound)
	} else if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	enodebState := st.ReportedState.(*lte_models.EnodebState)
	enodebState.TimeReported = st.TimeMs
	ent, err := configurator.LoadEntityForPhysicalID(st.ReporterID, configurator.EntityLoadCriteria{})
	if err == nil {
		enodebState.ReportingGatewayID = ent.Key
	}
	return c.JSON(http.StatusOK, enodebState)
}

func getNetworkAndEnbIDs(c echo.Context) (string, string, *echo.HTTPError) {
	vals, err := obsidian.GetParamValues(c, "network_id", "enodeb_serial")
	if err != nil {
		return "", "", err
	}
	return vals[0], vals[1], nil
}

func deleteConnectedEnodeb(c echo.Context) error {
	networkID, gatewayID, nerr := obsidian.GetNetworkAndGatewayIDs(c)
	if nerr != nil {
		return nerr
	}

	var enodebSerial string
	if err := c.Bind(&enodebSerial); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}

	_, err := configurator.UpdateEntity(networkID, (&lte_models.EnodebSerials{}).ToDeleteUpdateCriteria(networkID, gatewayID, enodebSerial))
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func addConnectedEnodeb(c echo.Context) error {
	networkID, gatewayID, nerr := obsidian.GetNetworkAndGatewayIDs(c)
	if nerr != nil {
		return nerr
	}

	var enodebSerial string
	if err := c.Bind(&enodebSerial); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}

	_, err := configurator.UpdateEntity(networkID, (&lte_models.EnodebSerials{}).ToCreateUpdateCriteria(networkID, gatewayID, enodebSerial))
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func listApns(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}

	ents, err := configurator.LoadAllEntitiesInNetwork(networkID, lte.APNEntityType, configurator.EntityLoadCriteria{LoadConfig: true})
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	ret := make(map[string]*lte_models.Apn, len(ents))
	for _, ent := range ents {
		ret[ent.Key] = (&lte_models.Apn{}).FromBackendModels(ent)
	}
	return c.JSON(http.StatusOK, ret)
}

func createApn(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}

	payload := &lte_models.Apn{}
	if err := c.Bind(payload); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if err := payload.ValidateModel(); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}

	_, err := configurator.CreateEntity(networkID, configurator.NetworkEntity{
		Type:   lte.APNEntityType,
		Key:    string(payload.ApnName),
		Config: payload.ApnConfiguration,
	})
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}

func getApnConfiguration(c echo.Context) error {
	networkID, apnName, nerr := getNetworkAndApnName(c)
	if nerr != nil {
		return nerr
	}

	ent, err := configurator.LoadEntity(networkID, lte.APNEntityType, apnName, configurator.EntityLoadCriteria{LoadConfig: true})
	switch {
	case err == merrors.ErrNotFound:
		return echo.ErrNotFound
	case err != nil:
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	ret := (&lte_models.Apn{}).FromBackendModels(ent)
	return c.JSON(http.StatusOK, ret)
}

func updateApnConfiguration(c echo.Context) error {
	networkID, apnName, nerr := getNetworkAndApnName(c)
	if nerr != nil {
		return nerr
	}

	payload := &lte_models.Apn{}
	if err := c.Bind(payload); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}
	if err := payload.ValidateModel(); err != nil {
		return obsidian.HttpError(err, http.StatusBadRequest)
	}

	_, err := configurator.LoadEntity(networkID, lte.APNEntityType, apnName, configurator.EntityLoadCriteria{})
	switch {
	case err == merrors.ErrNotFound:
		return echo.ErrNotFound
	case err != nil:
		return obsidian.HttpError(errors.Wrap(err, "failed to load existing APN"), http.StatusInternalServerError)
	}

	err = configurator.CreateOrUpdateEntityConfig(networkID, lte.APNEntityType, apnName, payload.ApnConfiguration)
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func deleteApnConfiguration(c echo.Context) error {
	networkID, apnName, nerr := getNetworkAndApnName(c)
	if nerr != nil {
		return nerr
	}

	ent, err := configurator.LoadEntity(networkID, lte.APNEntityType, apnName, configurator.EntityLoadCriteria{LoadAssocsToThis: true})
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	// Cascade deletes to all associated apn_resource and apn_policy_profile
	var deletes []storage.TypeAndKey
	deletes = append(deletes, ent.ParentAssociations.MultiFilter(lte.APNResourceEntityType, lte.APNPolicyProfileEntityType)...)
	deletes = append(deletes, ent.GetTypeAndKey())

	err = configurator.DeleteEntities(networkID, deletes)
	if err != nil {
		return obsidian.HttpError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func AddNetworkWideSubscriberRuleName(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}
	params, nerr := obsidian.GetParamValues(c, "rule_id")
	if nerr != nil {
		return nerr
	}
	err := addToNetworkSubscriberConfig(networkID, params[0], "")
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "Failed to update config"), http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

func AddNetworkWideSubscriberBaseName(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}
	params, nerr := obsidian.GetParamValues(c, "base_name")
	if nerr != nil {
		return nerr
	}
	err := addToNetworkSubscriberConfig(networkID, "", params[0])
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "Failed to update config"), http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

func RemoveNetworkWideSubscriberRuleName(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}
	params, nerr := obsidian.GetParamValues(c, "rule_id")
	if nerr != nil {
		return nerr
	}
	err := removeFromNetworkSubscriberConfig(networkID, params[0], "")
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "Failed to update config"), http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func RemoveNetworkWideSubscriberBaseName(c echo.Context) error {
	networkID, nerr := obsidian.GetNetworkId(c)
	if nerr != nil {
		return nerr
	}
	params, nerr := obsidian.GetParamValues(c, "base_name")
	if nerr != nil {
		return nerr
	}
	err := removeFromNetworkSubscriberConfig(networkID, "", params[0])
	if err != nil {
		return obsidian.HttpError(errors.Wrap(err, "Failed to update config"), http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func addToNetworkSubscriberConfig(networkID, ruleName, baseName string) error {
	network, err := configurator.LoadNetwork(networkID, false, true)
	if err != nil {
		return err
	}
	iSubscriberConfig, exists := network.Configs[lte.NetworkSubscriberConfigType]
	if !exists || iSubscriberConfig == nil {
		network.Configs[lte.NetworkSubscriberConfigType] = &policydb_models.NetworkSubscriberConfig{}
	}
	subscriberConfig, ok := network.Configs[lte.NetworkSubscriberConfigType].(*policydb_models.NetworkSubscriberConfig)
	if !ok {
		return fmt.Errorf("unable to convert config %v", subscriberConfig)
	}
	if len(ruleName) != 0 {
		ruleAlreadyExists := false
		for _, existing := range subscriberConfig.NetworkWideRuleNames {
			if existing == ruleName {
				ruleAlreadyExists = true
				break
			}
		}
		if !ruleAlreadyExists {
			subscriberConfig.NetworkWideRuleNames = append(subscriberConfig.NetworkWideRuleNames, ruleName)
		}
	}
	if len(baseName) != 0 {
		bnAlreadyExists := false
		for _, existing := range subscriberConfig.NetworkWideBaseNames {
			if existing == policydb_models.BaseName(baseName) {
				bnAlreadyExists = true
				break
			}
		}
		if !bnAlreadyExists {
			subscriberConfig.NetworkWideBaseNames = append(subscriberConfig.NetworkWideBaseNames, policydb_models.BaseName(baseName))
		}
	}
	return configurator.UpdateNetworkConfig(networkID, lte.NetworkSubscriberConfigType, subscriberConfig)
}

func removeFromNetworkSubscriberConfig(networkID, ruleName, baseName string) error {
	network, err := configurator.LoadNetwork(networkID, false, true)
	if err != nil {
		return err
	}
	iSubscriberConfig, exists := network.Configs[lte.NetworkSubscriberConfigType]
	if !exists || iSubscriberConfig == nil {
		network.Configs[lte.NetworkSubscriberConfigType] = &policydb_models.NetworkSubscriberConfig{}
	}
	subscriberConfig, ok := network.Configs[lte.NetworkSubscriberConfigType].(*policydb_models.NetworkSubscriberConfig)
	if !ok {
		return fmt.Errorf("unable to convert config")
	}
	if len(ruleName) != 0 {
		subscriberConfig.NetworkWideRuleNames = funk.FilterString(subscriberConfig.NetworkWideRuleNames,
			func(s string) bool { return s != ruleName })
	}
	if len(baseName) != 0 {
		subscriberConfig.NetworkWideBaseNames = funk.Filter(subscriberConfig.NetworkWideBaseNames,
			func(b policydb_models.BaseName) bool { return string(b) != baseName }).([]policydb_models.BaseName)
	}
	return configurator.UpdateNetworkConfig(networkID, lte.NetworkSubscriberConfigType, subscriberConfig)
}

func getNetworkAndApnName(c echo.Context) (string, string, *echo.HTTPError) {
	vals, err := obsidian.GetParamValues(c, "network_id", "apn_name")
	if err != nil {
		return "", "", err
	}
	return vals[0], vals[1], nil
}

func makeErr(err error) *echo.HTTPError {
	if err == merrors.ErrNotFound {
		return echo.ErrNotFound
	}
	return obsidian.HttpError(err, http.StatusInternalServerError)
}
