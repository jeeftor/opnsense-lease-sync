<?php

/**
 *    Copyright (C) 2024 DHCP AdGuard Sync
 *    All rights reserved.
 *
 *    Redistribution and use in source and binary forms, with or without
 *    modification, are permitted provided that the following conditions are met:
 *
 *    1. Redistributions of source code must retain the above copyright notice,
 *       this list of conditions and the following disclaimer.
 *
 *    2. Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 *    THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES,
 *    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
 *    AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
 *    AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 *    OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *    SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *    INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *    CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *    ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *    POSSIBILITY OF SUCH DAMAGE.
 */

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

            // Set form data - use dhcpadguardsync as the form name
            $mdl->setNodes($this->request->getPost("dhcpadguardsync"));

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
     * Get configuration
     */
    public function getAction()
    {
        $result = array();
        if ($this->request->isGet()) {
            $mdl = $this->getModel();
            $result['dhcpadguardsync'] = $mdl->getNodes();
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
