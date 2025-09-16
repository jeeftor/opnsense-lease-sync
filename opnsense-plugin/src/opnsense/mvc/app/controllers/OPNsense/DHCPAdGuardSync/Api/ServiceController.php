<?php
namespace OPNsense\DHCPAdGuardSync\Api;

use OPNsense\Base\ApiMutableServiceControllerBase;
use OPNsense\DHCPAdGuardSync\DHCPAdGuardSync;
use OPNsense\Core\Backend;

class ServiceController extends ApiMutableServiceControllerBase
{
    protected static $internalServiceClass = '\OPNsense\DHCPAdGuardSync\DHCPAdGuardSync';
    protected static $internalServiceTemplate = 'OPNsense/DHCPAdGuardSync';
    protected static $internalServiceEnabled = 'general.enabled';
    protected static $internalServiceName = 'dhcpadguardsync';

    /**
     * Additional custom action for testing configuration
     */
    public function testAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            $backend = new Backend();
            $response = $backend->configdRun("dhcpadguardsync test");
            if (strpos($response, "OK") !== false || strpos($response, "success") !== false) {
                $result['result'] = 'ok';
            }
            $result['response'] = $response;
        }
        return $result;
    }

    /**
     * Get service logs
     */
    public function logsAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            $backend = new Backend();
            $response = $backend->configdRun("dhcpadguardsync logs");
            $result['response'] = $response;
            $result['result'] = 'ok';
        }
        return $result;
    }
}
