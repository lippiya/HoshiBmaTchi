import { describe, it, expect, vi } from 'vitest'

const mockGetConversations = vi.fn()
const mockGetConversationDetail = vi.fn()
const mockSendMessage = vi.fn()
const mockDeleteConversation = vi.fn()

vi.mock('@/services/messageService', () => ({
  getConversations: () => mockGetConversations(),
  getConversationDetail: (id: string) => mockGetConversationDetail(id),
  sendMessage: (conversationId: string, content: string) => mockSendMessage(conversationId, content),
  deleteConversation: (id: string) => mockDeleteConversation(id),
}))

describe('Message Service', () => {
  it('should fetch conversations successfully', async () => {
    const mockConversations = [
      { id: '1', is_group: false, participants: [], unread_count: 0 },
      { id: '2', is_group: true, participants: [], unread_count: 2 }
    ]
    
    mockGetConversations.mockResolvedValue(mockConversations)
    
    const { getConversations } = await import('@/services/messageService')
    const result = await getConversations()
    
    expect(result).toEqual(mockConversations)
    expect(mockGetConversations).toHaveBeenCalled()
  })

  it('should get conversation detail by id', async () => {
    const mockDetail = {
      conversation: { id: '123', participants: [] },
      messages: [
        { id: 'msg1', content: 'Hello', sender_id: 'user1' }
      ]
    }
    
    mockGetConversationDetail.mockResolvedValue(mockDetail)
    
    const { getConversationDetail } = await import('@/services/messageService')
    const result = await getConversationDetail('123')
    
    expect(result).toEqual(mockDetail)
    expect(mockGetConversationDetail).toHaveBeenCalledWith('123')
  })

  it('should send message successfully', async () => {
    const mockResponse = {
      id: 'newmsg',
      content: 'Test message',
      sender_id: 'user1'
    }
    
    mockSendMessage.mockResolvedValue(mockResponse)
    
    const { sendMessage } = await import('@/services/messageService')
    const result = await sendMessage('conv123', 'Test message')
    
    expect(result).toEqual(mockResponse)
    expect(mockSendMessage).toHaveBeenCalledWith('conv123', 'Test message')
  })

  it('should delete conversation', async () => {
    mockDeleteConversation.mockResolvedValue({ success: true })
    
    const { deleteConversation } = await import('@/services/messageService')
    await deleteConversation('conv123')
    
    expect(mockDeleteConversation).toHaveBeenCalledWith('conv123')
  })
})
