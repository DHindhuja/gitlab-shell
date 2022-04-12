require_relative 'spec_helper'

require 'open3'
require 'json'

describe 'bin/gitlab-shell 2fa_verify push' do
  include_context 'gitlab shell'

  let(:env) do
    { 'SSH_CONNECTION' => 'fake',
      'SSH_ORIGINAL_COMMAND' => '2fa_verify' }
  end

  before(:context) do
    write_config('gitlab_url' => "http+unix://#{CGI.escape(tmp_socket_path)}")
  end

  def mock_server(server)
    server.mount_proc('/api/v4/internal/two_factor_push_otp_check') do |req, res|
      res.content_type = 'application/json'
      res.status = 200

      params = JSON.parse(req.body)
      key_id = params['key_id'] || params['user_id'].to_s

      if key_id == '100'
        res.body = { success: false }.to_json
      elsif key_id == '102'
        res.body = { success: true }.to_json
      else
        res.body = { success: false, message: 'boom!' }.to_json
      end
    end

    server.mount_proc('/api/v4/internal/discover') do |_, res|
      res.status = 200
      res.content_type = 'application/json'
      res.body = { id: 100, name: 'Some User', username: 'someuser' }.to_json
    end
  end

  describe 'command' do
    context 'when push is provided' do
      let(:cmd) { "#{gitlab_shell_path} key-102" }

      it 'prints a successful push verification message' do
        verify_successful_verification_push!(cmd)
      end
    end

    context 'when API error occurs' do
      let(:cmd) { "#{gitlab_shell_path} key-101" }

      it 'prints the error message' do
        Open3.popen2(env, cmd) do |stdin, stdout|
          expect(stdout.gets(5)).to eq('OTP: ')

          expect(stdout.flush.read).to eq("\nPush OTP validation failed.\nboom!\n")
        end
      end
    end
  end

  def verify_successful_verification_push!(cmd)
    Open3.popen2(env, cmd) do |stdin, stdout|
      expect(stdout.gets(5)).to eq('OTP: ')

      expect(stdout.flush.read).to eq("\nPush OTP validation successful. Git operations are now allowed.\n")
    end
  end
end
