<?php
namespace OPNsense\DHCPAdGuardSync\Api;

use OPNsense\Base\ApiMutableModelControllerBase;
use OPNsense\Core\Config;
use OPNsense\DHCPAdGuardSync\DHCPAdGuardSync;

class SettingsController extends ApiMutableModelControllerBase
{
    protected static $internalModelName = 'dhcpadguardsync';
    protected static $internalModelClass = '\OPNsense\DHCPAdGuardSync\DHCPAdGuardSync';

    public function getAction()
    {
        $result = array();
        if ($this->request->isGet()) {
            $mdl = new DHCPAdGuardSync();
            $result['dhcpadguardsync'] = $mdl->getNodes();
        }
        return $result;
    }

    public function setAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            $mdl = new DHCPAdGuardSync();
            $mdl->setNodes($this->request->getPost("dhcpadguardsync"));
            $validationMessages = $mdl->performValidation();

            if (count($validationMessages) == 0) {
                $mdl->serializeToConfig();
                Config::getInstance()->save();
                $result["result"] = "saved";
            } else {
                $result["validations"] = $validationMessages;
            }
        }
        return $result;
    }
}
