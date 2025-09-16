<?php
namespace OPNsense\DHCPAdGuardSync\Api;

use OPNsense\Base\ApiMutableModelControllerBase;
use OPNsense\Core\Config;
use OPNsense\DHCPAdGuardSync\DHCPAdGuardSync;

class SettingsController extends ApiMutableModelControllerBase
{
    protected static $internalModelClass = '\OPNsense\DHCPAdGuardSync\DHCPAdGuardSync';
    protected static $internalModelName = 'dhcpadguardsync';

    /**
     * Override the set action to handle configuration generation
     */
    public function setAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            // Get the model
            $mdl = $this->getModel();

            // Set form data
            $mdl->setNodes($this->request->getPost("frm_GeneralSettings"));

            // Perform validation
            $validationMessages = $mdl->performValidation();
            if (count($validationMessages) == 0) {
                // Save to config.xml
                $mdl->serializeToConfig();
                Config::getInstance()->save();

                // Generate configuration file
                $this->generateConfigFile($mdl);

                $result["result"] = "saved";
            } else {
                $result["validations"] = $validationMessages;
            }
        }
        return $result;
    }

    /**
     * Get configuration - use proper form name
     */
    public function getAction()
    {
        $result = array();
        if ($this->request->isGet()) {
            $mdl = $this->getModel();
            $result['frm_GeneralSettings'] = $mdl->getNodes();
        }
        return $result;
    }

    private function generateConfigFile($model)
    {
        $config = $model->getNodes();
        $general = $config['general'];

        $configContent = "# DHCP AdGuard Sync Configuration\n";
        $configContent .= "ADGUARD_USERNAME=\"" . (string)$general['adguard_username'] . "\"\n";
        $configContent .= "ADGUARD_PASSWORD=\"" . (string)$general['adguard_password'] . "\"\n";
        $configContent .= "ADGUARD_URL=\"" . (string)$general['adguard_url'] . "\"\n";
        $configContent .= "ADGUARD_SCHEME=\"http\"\n";

        if ((string)$general['dhcp_server'] === 'dnsmasq') {
            $configContent .= "DHCP_LEASE_PATH=\"/var/db/dnsmasq.leases\"\n";
            $configContent .= "LEASE_FORMAT=\"dnsmasq\"\n";
        } else {
            $configContent .= "DHCP_LEASE_PATH=\"/var/dhcpd/var/db/dhcpd.leases\"\n";
            $configContent .= "LEASE_FORMAT=\"isc\"\n";
        }

        $configContent .= "LOG_LEVEL=\"info\"\n";
        $configContent .= "LOG_FILE=\"/var/log/dhcp-adguard-sync.log\"\n";

        // Write config file
        $configDir = '/usr/local/etc/dhcp-adguard-sync';
        if (!is_dir($configDir)) {
            mkdir($configDir, 0755, true);
        }

        file_put_contents($configDir . '/config.yaml', $configContent);
    }
}
