<?php

/*
 * Copyright (C) 2023 Deciso B.V.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES,
 * INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
 * AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
 * AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 * OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

namespace OPNsense\Dhcpsync\Api;

use OPNsense\Base\ApiMutableModelControllerBase;
use OPNsense\Core\Config;
use OPNsense\Dhcpsync\Dhcpsync;

/**
 * Class GeneralController
 * @package OPNsense\Dhcpsync\Api
 */
class GeneralController extends ApiMutableModelControllerBase
{
    protected static $internalModelClass = '\\OPNsense\\Dhcpsync\\Dhcpsync';
    protected static $internalModelName = 'dhcpsync';
    protected $debug = false; // Set to true to enable debug logging

    private function debugLog($message, $data = null)
    {
        if ($this->debug) {
            $logMessage = "[Dhcpsync Debug] {$message}";
            if ($data !== null) {
                $logMessage .= ": " . print_r($data, true);
            }
            error_log($logMessage);
        }
    }

    protected function sessionClose()
    {
        $this->debugLog('Session closed for async operation');
    }

    private $configFile = '/usr/local/etc/dhcpsync/config.env';

    private function readConfigFile()
    {
        $config = [];
        if (file_exists($this->configFile)) {
            try {
                $fileContent = file_get_contents($this->configFile);
                if ($fileContent === false) {
                    $this->debugLog("Failed to read config file {$this->configFile}");
                    return [];
                }
                $config = $this->parseYamlContent($fileContent);
                $this->debugLog('Config file contents', $config);
                return $config;
            } catch (\Exception $e) {
                $this->debugLog("Error reading config file {$this->configFile}", $e->getMessage());
            }
        } else {
            $this->debugLog("Config file does not exist: {$this->configFile}");
        }
        return $config;
    }

    private function parseYamlContent($content)
    {
        $config = [];
        $lines = explode("\n", $content);

        foreach ($lines as $line) {
            $line = trim($line);
            if (empty($line) || $line[0] === '#') {
                continue;
            }

            if (strpos($line, '=') !== false) {
                list($key, $value) = explode('=', $line, 2);
                $key = trim($key);

                $isCommented = false;
                if (substr($key, 0, 1) === '#') {
                    $isCommented = true;
                    $key = ltrim($key, '#');
                }

                $value = trim($value, " \"'");

                if ($isCommented) {
                    $config['#' . $key] = $value;
                } else {
                    $config[$key] = $value;
                }
            }
        }

        return $config;
    }

    private function writeConfigFile($config)
    {
        try {
            $configDir = dirname($this->configFile);
            if (!file_exists($configDir)) {
                if (!mkdir($configDir, 0755, true)) {
                    throw new \Exception("Failed to create config directory: {$configDir}");
                }
            }

            $sections = [
                '# AdGuard Home credentials' => ['ADGUARD_USERNAME', 'ADGUARD_PASSWORD'],
                '# AdGuard Home connection settings' => ['ADGUARD_URL', 'ADGUARD_SCHEME'],
                '# DHCP lease file configuration' => ['DHCP_LEASE_PATH', 'LEASE_FORMAT'],
                '# Optional settings' => ['PRESERVE_DELETED_HOSTS', 'DEBUG', 'DRY_RUN', 'ADGUARD_TIMEOUT'],
                '# Logging configuration - OPNsense optimized' => ['LOG_LEVEL', 'LOG_FILE', 'SYSLOG_FACILITY', 'MAX_LOG_SIZE', 'MAX_BACKUPS', 'MAX_AGE', 'NO_COMPRESS']
            ];

            $defaults = [
                'ADGUARD_USERNAME' => '',
                'ADGUARD_PASSWORD' => '',
                'ADGUARD_URL' => 'localhost:3000',
                'ADGUARD_SCHEME' => 'http',
                'DHCP_LEASE_PATH' => '/var/db/dnsmasq.leases',
                'LEASE_FORMAT' => 'dnsmasq',
                'PRESERVE_DELETED_HOSTS' => 'false',
                'DEBUG' => 'false',
                'DRY_RUN' => 'false',
                'ADGUARD_TIMEOUT' => '10',
                'LOG_LEVEL' => 'info',
                'LOG_FILE' => '/var/log/dhcpsync.log',
                'SYSLOG_FACILITY' => 'local3',
                'MAX_LOG_SIZE' => '100',
                'MAX_BACKUPS' => '3',
                'MAX_AGE' => '28',
                'NO_COMPRESS' => 'false'
            ];

            $commentedByDefault = ['PRESERVE_DELETED_HOSTS', 'DEBUG', 'DRY_RUN', 'NO_COMPRESS'];

            $content = "";

            foreach ($sections as $sectionHeader => $fields) {
                $content .= $sectionHeader . "\n";

                foreach ($fields as $field) {
                    $isCommented = false;

                    if (isset($config[$field])) {
                        $value = $config[$field];
                    } elseif (isset($config['#' . $field])) {
                        $value = $config['#' . $field];
                        $isCommented = true;
                    } else {
                        $value = $defaults[$field];
                        $isCommented = in_array($field, $commentedByDefault);
                    }

                    $prefix = $isCommented ? '#' : '';
                    $content .= $prefix . $field . "=\"" . $value . "\"";

                    if ($field === 'LEASE_FORMAT') {
                        $content .= "    # Lease format: \"isc\" or \"dnsmasq\"";
                    }

                    $content .= "\n";
                }

                $content .= "\n";
            }

            $result = file_put_contents($this->configFile, $content);
            if ($result === false) {
                throw new \Exception("Failed to write to config file: {$this->configFile}");
            }

            return true;
        } catch (\Exception $e) {
            error_log("Error writing config file {$this->configFile}: " . $e->getMessage());
            return false;
        }
    }

    private function configToModel($config)
    {
        $this->debugLog('configToModel called with config', $config);
        $modelConfig = [];

        try {
            if (isset($config['ADGUARD_USERNAME'])) {
                $modelConfig['AdguardHomeUsername'] = trim($config['ADGUARD_USERNAME'], '"\' ');
            }

            if (isset($config['ADGUARD_PASSWORD'])) {
                $modelConfig['AdguardHomePassword'] = trim($config['ADGUARD_PASSWORD'], '"\' ');
            }

            if (isset($config['ADGUARD_URL']) && isset($config['ADGUARD_SCHEME'])) {
                $scheme = trim($config['ADGUARD_SCHEME'], '"\' ');
                $url = trim($config['ADGUARD_URL'], '"\' ');
                $modelConfig['AdguardHomeURL'] = $scheme . '://' . $url;
            }

            if (isset($config['LEASE_FORMAT'])) {
                $format = $config['LEASE_FORMAT'];
                if (strpos($format, '#') !== false) {
                    $format = trim(substr($format, 0, strpos($format, '#')));
                }
                $format = trim($format, '"\' ');
                $format = strtolower(trim($format));
                if ($format !== 'dnsmasq' && $format !== 'isc') {
                    $format = 'dnsmasq';
                }
                $modelConfig['DHCPServer'] = $format;
            }

            if (!isset($modelConfig['AdguardHomeUsername'])) {
                $modelConfig['AdguardHomeUsername'] = '';
            }
            if (!isset($modelConfig['AdguardHomePassword'])) {
                $modelConfig['AdguardHomePassword'] = '';
            }
            if (!isset($modelConfig['AdguardHomeURL'])) {
                $modelConfig['AdguardHomeURL'] = 'http://localhost:3000';
            }
            if (!isset($modelConfig['DHCPServer'])) {
                $modelConfig['DHCPServer'] = 'dnsmasq';
            }

        } catch (\Exception $e) {
            $this->debugLog('Error converting config to model', $e->getMessage());
        }

        return $modelConfig;
    }

    private function modelToConfig($modelConfig, $fullConfig = [])
    {
        try {
            $config = !empty($fullConfig) ? $fullConfig : $this->readConfigFile();

            if (isset($modelConfig['AdguardHomeUsername'])) {
                $config['ADGUARD_USERNAME'] = $modelConfig['AdguardHomeUsername'];
            }
            if (isset($modelConfig['AdguardHomePassword'])) {
                $config['ADGUARD_PASSWORD'] = $modelConfig['AdguardHomePassword'];
            }
            if (isset($modelConfig['AdguardHomeURL'])) {
                $url = $modelConfig['AdguardHomeURL'];
                if (!preg_match('/^https?:\\/\\//i', $url)) {
                    $url = 'http://' . $url;
                }
                $parsedUrl = parse_url($url);
                if (isset($parsedUrl['host'])) {
                    $config['ADGUARD_URL'] = $parsedUrl['host'];
                    if (isset($parsedUrl['port'])) {
                        $config['ADGUARD_URL'] .= ':' . $parsedUrl['port'];
                    }
                } else {
                    $config['ADGUARD_URL'] = 'localhost:3000';
                }
                $config['ADGUARD_SCHEME'] = isset($parsedUrl['scheme']) ? $parsedUrl['scheme'] : 'http';
            }
            if (isset($modelConfig['DHCPServer'])) {
                if (is_array($modelConfig['DHCPServer'])) {
                    $format = 'dnsmasq';
                    foreach ($modelConfig['DHCPServer'] as $key => $option) {
                        if (isset($option['selected']) && $option['selected'] == 1) {
                            $format = $key;
                            break;
                        }
                    }
                } else {
                    $format = strtolower(trim($modelConfig['DHCPServer']));
                }
                if ($format !== 'dnsmasq' && $format !== 'isc') {
                    $format = 'dnsmasq';
                }
                $config['LEASE_FORMAT'] = $format;
                if ($format === 'dnsmasq') {
                    $config['DHCP_LEASE_PATH'] = '/var/db/dnsmasq.leases';
                } else if ($format === 'isc') {
                    $config['DHCP_LEASE_PATH'] = '/var/dhcpd/var/db/dhcpd.leases';
                } else {
                    $config['DHCP_LEASE_PATH'] = '/var/db/dnsmasq.leases';
                }
            }

            $defaults = [
                'ADGUARD_USERNAME' => '',
                'ADGUARD_PASSWORD' => '',
                'ADGUARD_URL' => 'localhost:3000',
                'ADGUARD_SCHEME' => 'http',
                'DHCP_LEASE_PATH' => '/var/db/dnsmasq.leases',
                'LEASE_FORMAT' => 'dnsmasq',
                'ADGUARD_TIMEOUT' => '10',
                'LOG_LEVEL' => 'info',
                'LOG_FILE' => '/var/log/dhcpsync.log',
                'SYSLOG_FACILITY' => 'local3',
                'MAX_LOG_SIZE' => '100',
                'MAX_BACKUPS' => '3',
                'MAX_AGE' => '28'
            ];

            foreach ($defaults as $key => $value) {
                if (!isset($config[$key])) {
                    $config[$key] = $value;
                }
            }

            return $config;
        } catch (\Exception $e) {
            error_log("Error converting model to config: " . $e->getMessage());
            return [];
        }
    }

    public function getAction()
    {
        try {
            $result = parent::getAction();
            $config = $this->readConfigFile();
            $modelConfig = $this->configToModel($config);

            if (isset($result['dhcpsync']) && isset($result['dhcpsync']['general'])) {
                foreach ($modelConfig as $key => $value) {
                    if (isset($result['dhcpsync']['general'][$key])) {
                        if ($key === 'Enabled') {
                            continue;
                        }
                        else if ($key === 'DHCPServer') {
                            if (isset($result['dhcpsync']['general'][$key][$value])) {
                                $result['dhcpsync']['general'][$key][$value]['selected'] = 1;
                            } else {
                                $result['dhcpsync']['general'][$key] = $value;
                            }
                        } else {
                            $result['dhcpsync']['general'][$key] = $value;
                        }
                    }
                }
            }

            return $result;
        } catch (\Exception $e) {
            $this->debugLog('Error in getAction', $e->getMessage());
            throw $e;
        }
    }

    public function setAction()
    {
        try {
            $result = parent::setAction();

            if ($result['result'] === 'saved') {
                $mdl = new Dhcpsync();
                $nodes = $mdl->getNodes();

                if (isset($nodes['general'])) {
                    $modelConfig = $nodes['general'];
                    $config = $this->modelToConfig($modelConfig);
                    if (!$this->writeConfigFile($config)) {
                        $result['result'] = 'failed';
                        $result['validations'] = array('Failed to write config file');
                    }
                }
            }

            return $result;
        } catch (\Exception $e) {
            return [
                'result' => 'failed',
                'error' => $e->getMessage(),
                'exception' => get_class($e)
            ];
        }
    }

    public function testConnectionAction()
    {
        $result = array('status' => 'failed');

        if ($this->request->isPost()) {
            $this->sessionClose();

            $mdl = new Dhcpsync();

            $url = $mdl->general->AdguardHomeURL->__toString();
            $username = $mdl->general->AdguardHomeUsername->__toString();
            $password = $mdl->general->AdguardHomePassword->__toString();

            if (empty($url) || empty($username)) {
                $result['message'] = 'Please fill in all required fields (URL, Username)';
                return $result;
            }

            $url = rtrim($url, '/');

            if (preg_match('/\/api$/', $url)) {
                $testUrl = $url . '/control/status';
            }
            else if (preg_match('/\/control$/', $url)) {
                $testUrl = $url . '/status';
            }
            else {
                $testUrl = $url . '/control/status';
            }

            $ch = curl_init();

            curl_setopt($ch, CURLOPT_URL, $testUrl);
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HEADER, true);
            curl_setopt($ch, CURLOPT_USERPWD, $username . ':' . $password);
            curl_setopt($ch, CURLOPT_TIMEOUT, 30);
            curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 10);
            curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
            curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);

            $verbose = fopen('php://temp', 'w+');
            curl_setopt($ch, CURLOPT_STDERR, $verbose);
            curl_setopt($ch, CURLOPT_VERBOSE, true);

            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            $error = curl_error($ch);

            rewind($verbose);
            $verboseLog = stream_get_contents($verbose);
            fclose($verbose);

            $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
            $headers = substr($response, 0, $headerSize);
            $body = substr($response, $headerSize);

            curl_close($ch);

            $result['debug'] = [
                'url' => $testUrl,
                'http_code' => $httpCode,
                'error' => $error,
                'headers' => $headers,
                'response_preview' => substr($body, 0, 500)
            ];

            if ($httpCode === 200) {
                $result['status'] = 'ok';
                $result['message'] = 'Connection successful!';
            } elseif ($httpCode === 401 || $httpCode === 403) {
                $result['message'] = 'Authentication failed. Please check your username and password.';
            } else {
                $result['message'] = 'Connection failed: ' . ($error ? $error : 'HTTP Error ' . $httpCode);
            }
        }

        return $result;
    }
}
