class Order < ApplicationRecord
  belongs_to :user
  has_many :order_items, dependent: :destroy
  has_one :shipping_address, dependent: :destroy
  has_many :payments, dependent: :destroy
  has_many :refunds, dependent: :destroy

  validates :status, presence: true, inclusion: { in: %w[pending confirmed shipped delivered cancelled refunded] }
  validates :total_price, numericality: { greater_than_or_equal_to: 0 }
  validate :status_transition_validity, on: :update

  scope :recent, -> { order(created_at: :desc) }
  scope :by_status, ->(status) { where(status: status) }
  scope :by_date_range, ->(start_date, end_date) { where(created_at: start_date..end_date) }
  scope :high_value, -> { where('total_price > ?', 1000) }
  scope :pending_shipment, -> { where(status: 'confirmed') }

  before_create :set_default_status
  after_update :log_status_change, if: :saved_change_to_status?

  def subtotal
    order_items.sum { |item| item.price * item.quantity }
  end

  def tax_amount
    subtotal * 0.08
  end

  def total
    subtotal + tax_amount
  end

  def apply_coupon(discount_amount)
    self.total_price -= discount_amount
    save
  end

  def can_cancel?
    %w[pending confirmed].include?(status)
  end

  def ship!
    update!(status: 'shipped') if status == 'confirmed'
  end

  def complete!
    update!(status: 'delivered') if status == 'shipped'
  end

  def refund!(reason)
    transaction do
      refunds.create!(amount: total, reason: reason)
      update!(status: 'refunded')
    end
  end

  private

  def set_default_status
    self.status ||= 'pending'
  end

  def status_transition_validity
    return unless status_changed?

    previous_status = status_change_to_hash['status'][0]
    case previous_status
    when 'delivered'
      errors.add(:status, "cannot change status from delivered")
    when 'cancelled', 'refunded'
      errors.add(:status, "cannot change status from #{previous_status}")
    end
  end

  def log_status_change
    Rails.logger.info "Order #{id} transitioned to #{status}"
  end
end