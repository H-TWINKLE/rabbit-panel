import request from '@/utils/request'

export interface AgentConfig {
    api_url: string
    api_key: string
    model: string
    enabled: boolean
}

export function getAgentConfig() {
    return request.get<AgentConfig>('/settings/agent')
}

export function saveAgentConfig(data: AgentConfig) {
    return request.post('/settings/agent', data)
}
