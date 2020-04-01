package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsElasticTranscoderPreset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticTranscoderPresetCreate,
		Read:   resourceAwsElasticTranscoderPresetRead,
		Delete: resourceAwsElasticTranscoderPresetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"audio": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.AudioParameters
					Schema: map[string]*schema.Schema{
						"audio_packing_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bit_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"channels": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"codec": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sample_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"audio_codec_options": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bit_depth": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bit_order": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"profile": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"signed": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"container": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"thumbnails": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					// elastictranscoder.Thumbnails
					Schema: map[string]*schema.Schema{
						"aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"interval": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"padding_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"resolution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"video": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.VideoParameters
					Schema: map[string]*schema.Schema{
						"aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bit_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"codec": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"display_aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"fixed_gop": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"frame_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"keyframes_max_dist": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_frame_rate": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "30",
							ForceNew: true,
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"padding_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"resolution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Default:  "Fit",
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"video_watermarks": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					// elastictranscoder.PresetWatermark
					Schema: map[string]*schema.Schema{
						"horizontal_align": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"horizontal_offset": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"opacity": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"vertical_align": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"vertical_offset": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"video_codec_options": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsElasticTranscoderPresetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	req := &elastictranscoder.CreatePresetInput{
		Audio:       expandETAudioParams(d),
		Container:   aws.String(d.Get("container").(string)),
		Description: aws.String(d.Get("description").(string)),
		Thumbnails:  expandETThumbnails(d),
		Video:       expandETVideoParams(d),
	}

	if name, ok := d.GetOk("name"); ok {
		req.Name = aws.String(name.(string))
	} else {
		name := resource.PrefixedUniqueId("tf-et-preset-")
		d.Set("name", name)
		req.Name = aws.String(name)
	}

	log.Printf("[DEBUG] Elastic Transcoder Preset create opts: %s", req)
	resp, err := conn.CreatePreset(req)
	if err != nil {
		return fmt.Errorf("Error creating Elastic Transcoder Preset: %s", err)
	}

	if resp.Warning != nil && *resp.Warning != "" {
		log.Printf("[WARN] Elastic Transcoder Preset: %s", *resp.Warning)
	}

	d.SetId(*resp.Preset.Id)
	d.Set("arn", resp.Preset.Arn)

	return nil
}

func expandETThumbnails(d *schema.ResourceData) *elastictranscoder.Thumbnails {
	set, ok := d.GetOk("thumbnails")
	if !ok {
		return nil
	}

	s := set.(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}
	t := s.List()[0].(map[string]interface{})

	thumbnails := &elastictranscoder.Thumbnails{}

	if v, ok := t["aspect_ratio"]; ok && v.(string) != "" {
		thumbnails.AspectRatio = aws.String(v.(string))
	}

	if v, ok := t["interval"]; ok && v.(string) != "" {
		thumbnails.Interval = aws.String(v.(string))
	}

	if v, ok := t["format"]; ok && v.(string) != "" {
		thumbnails.Format = aws.String(v.(string))
	}

	if v, ok := t["max_height"]; ok && v.(string) != "" {
		thumbnails.MaxHeight = aws.String(v.(string))
	}

	if v, ok := t["max_width"]; ok && v.(string) != "" {
		thumbnails.MaxWidth = aws.String(v.(string))
	}

	if v, ok := t["padding_policy"]; ok && v.(string) != "" {
		thumbnails.PaddingPolicy = aws.String(v.(string))
	}

	if v, ok := t["resolution"]; ok && v.(string) != "" {
		thumbnails.Resolution = aws.String(v.(string))
	}

	if v, ok := t["sizing_policy"]; ok && v.(string) != "" {
		thumbnails.SizingPolicy = aws.String(v.(string))
	}

	return thumbnails
}

func expandETAudioParams(d *schema.ResourceData) *elastictranscoder.AudioParameters {
	set, ok := d.GetOk("audio")
	if !ok {
		return nil
	}

	s := set.(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}
	audio := s.List()[0].(map[string]interface{})

	return &elastictranscoder.AudioParameters{
		AudioPackingMode: aws.String(audio["audio_packing_mode"].(string)),
		BitRate:          aws.String(audio["bit_rate"].(string)),
		Channels:         aws.String(audio["channels"].(string)),
		Codec:            aws.String(audio["codec"].(string)),
		CodecOptions:     expandETAudioCodecOptions(d),
		SampleRate:       aws.String(audio["sample_rate"].(string)),
	}
}

func expandETAudioCodecOptions(d *schema.ResourceData) *elastictranscoder.AudioCodecOptions {
	s := d.Get("audio_codec_options").(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}

	codec := s.List()[0].(map[string]interface{})

	codecOpts := &elastictranscoder.AudioCodecOptions{}

	if v, ok := codec["signed"]; ok && v.(string) != "" {
		codecOpts.Signed = aws.String(v.(string))
	}

	if v, ok := codec["profile"]; ok && v.(string) != "" {
		codecOpts.Profile = aws.String(v.(string))
	}

	if v, ok := codec["bit_order"]; ok && v.(string) != "" {
		codecOpts.BitOrder = aws.String(v.(string))
	}

	if v, ok := codec["bit_depth"]; ok && v.(string) != "" {
		codecOpts.BitDepth = aws.String(v.(string))
	}

	return codecOpts
}

func expandETVideoParams(d *schema.ResourceData) *elastictranscoder.VideoParameters {
	s := d.Get("video").(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}
	p := s.List()[0].(map[string]interface{})

	etVideoParams := &elastictranscoder.VideoParameters{
		Watermarks: expandETVideoWatermarks(d),
	}

	if v, ok := d.GetOk("video_codec_options"); ok {
		codecOpts := make(map[string]string)
		for k, va := range v.(map[string]interface{}) {
			codecOpts[k] = va.(string)
		}

		etVideoParams.CodecOptions = aws.StringMap(codecOpts)
	}

	if v, ok := p["aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.AspectRatio = aws.String(v.(string))
	}

	if v, ok := p["bit_rate"]; ok && v.(string) != "" {
		etVideoParams.BitRate = aws.String(v.(string))
	}

	if v, ok := p["display_aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.DisplayAspectRatio = aws.String(v.(string))
	}

	if v, ok := p["aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.AspectRatio = aws.String(v.(string))
	}

	if v, ok := p["fixed_gop"]; ok && v.(string) != "" {
		etVideoParams.FixedGOP = aws.String(v.(string))
	}

	if v, ok := p["frame_rate"]; ok && v.(string) != "" {
		etVideoParams.FrameRate = aws.String(v.(string))
	}

	if v, ok := p["keyframes_max_dist"]; ok && v.(string) != "" {
		etVideoParams.KeyframesMaxDist = aws.String(v.(string))
	}

	if v, ok := p["max_frame_rate"]; ok && v.(string) != "" {
		etVideoParams.MaxFrameRate = aws.String(v.(string))
	}

	if v, ok := p["max_height"]; ok && v.(string) != "" {
		etVideoParams.MaxHeight = aws.String(v.(string))
	}

	if v, ok := p["max_width"]; ok && v.(string) != "" {
		etVideoParams.MaxWidth = aws.String(v.(string))
	}

	if v, ok := p["padding_policy"]; ok && v.(string) != "" {
		etVideoParams.PaddingPolicy = aws.String(v.(string))
	}

	if v, ok := p["resolution"]; ok && v.(string) != "" {
		etVideoParams.Resolution = aws.String(v.(string))
	}

	if v, ok := p["sizing_policy"]; ok && v.(string) != "" {
		etVideoParams.SizingPolicy = aws.String(v.(string))
	}

	if v, ok := p["codec"]; ok && v.(string) != "" {
		etVideoParams.Codec = aws.String(v.(string))
	}

	return etVideoParams
}

func expandETVideoWatermarks(d *schema.ResourceData) []*elastictranscoder.PresetWatermark {
	s := d.Get("video_watermarks").(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}
	var watermarks []*elastictranscoder.PresetWatermark

	for _, w := range s.List() {
		p := w.(map[string]interface{})
		watermark := &elastictranscoder.PresetWatermark{
			HorizontalAlign:  aws.String(p["horizontal_align"].(string)),
			HorizontalOffset: aws.String(p["horizontal_offset"].(string)),
			Id:               aws.String(p["id"].(string)),
			MaxHeight:        aws.String(p["max_height"].(string)),
			MaxWidth:         aws.String(p["max_width"].(string)),
			Opacity:          aws.String(p["opacity"].(string)),
			SizingPolicy:     aws.String(p["sizing_policy"].(string)),
			Target:           aws.String(p["target"].(string)),
			VerticalAlign:    aws.String(p["vertical_align"].(string)),
			VerticalOffset:   aws.String(p["vertical_offset"].(string)),
		}
		watermarks = append(watermarks, watermark)
	}

	return watermarks
}

func resourceAwsElasticTranscoderPresetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	resp, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Elastic Transcoder Preset Read response: %#v", resp)

	preset := resp.Preset
	d.Set("arn", preset.Arn)

	if preset.Audio != nil {
		err := d.Set("audio", flattenETAudioParameters(preset.Audio))
		if err != nil {
			return err
		}

		if preset.Audio.CodecOptions != nil {
			d.Set("audio_codec_options", flattenETAudioCodecOptions(preset.Audio.CodecOptions))
		}
	}

	d.Set("container", preset.Container)
	d.Set("name", preset.Name)
	d.Set("description", preset.Description)

	if preset.Thumbnails != nil {
		err := d.Set("thumbnails", flattenETThumbnails(preset.Thumbnails))
		if err != nil {
			return err
		}
	}

	d.Set("type", preset.Type)

	if preset.Video != nil {
		err := d.Set("video", flattenETVideoParams(preset.Video))
		if err != nil {
			return err
		}

		if preset.Video.CodecOptions != nil {
			d.Set("video_codec_options", aws.StringValueMap(preset.Video.CodecOptions))
		}

		if preset.Video.Watermarks != nil {
			d.Set("video_watermarks", flattenETWatermarks(preset.Video.Watermarks))
		}
	}

	return nil
}

func flattenETAudioParameters(audio *elastictranscoder.AudioParameters) []map[string]interface{} {
	m := setMap(make(map[string]interface{}))

	m.SetString("audio_packing_mode", audio.AudioPackingMode)
	m.SetString("bit_rate", audio.BitRate)
	m.SetString("channels", audio.Channels)
	m.SetString("codec", audio.Codec)
	m.SetString("sample_rate", audio.SampleRate)

	return m.MapList()
}

func flattenETAudioCodecOptions(opts *elastictranscoder.AudioCodecOptions) []map[string]interface{} {
	if opts == nil {
		return nil
	}

	m := setMap(make(map[string]interface{}))

	m.SetString("bit_depth", opts.BitDepth)
	m.SetString("bit_order", opts.BitOrder)
	m.SetString("profile", opts.Profile)
	m.SetString("signed", opts.Signed)

	return m.MapList()
}

func flattenETThumbnails(thumbs *elastictranscoder.Thumbnails) []map[string]interface{} {
	m := setMap(make(map[string]interface{}))

	m.SetString("aspect_ratio", thumbs.AspectRatio)
	m.SetString("format", thumbs.Format)
	m.SetString("interval", thumbs.Interval)
	m.SetString("max_height", thumbs.MaxHeight)
	m.SetString("max_width", thumbs.MaxWidth)
	m.SetString("padding_policy", thumbs.PaddingPolicy)
	m.SetString("resolution", thumbs.Resolution)
	m.SetString("sizing_policy", thumbs.SizingPolicy)

	return m.MapList()
}

func flattenETVideoParams(video *elastictranscoder.VideoParameters) []map[string]interface{} {
	m := setMap(make(map[string]interface{}))

	m.SetString("aspect_ratio", video.AspectRatio)
	m.SetString("bit_rate", video.BitRate)
	m.SetString("codec", video.Codec)
	m.SetString("display_aspect_ratio", video.DisplayAspectRatio)
	m.SetString("fixed_gop", video.FixedGOP)
	m.SetString("frame_rate", video.FrameRate)
	m.SetString("keyframes_max_dist", video.KeyframesMaxDist)
	m.SetString("max_frame_rate", video.MaxFrameRate)
	m.SetString("max_height", video.MaxHeight)
	m.SetString("max_width", video.MaxWidth)
	m.SetString("padding_policy", video.PaddingPolicy)
	m.SetString("resolution", video.Resolution)
	m.SetString("sizing_policy", video.SizingPolicy)

	return m.MapList()
}

func flattenETWatermarks(watermarks []*elastictranscoder.PresetWatermark) []map[string]interface{} {
	var watermarkSet []map[string]interface{}

	for _, w := range watermarks {
		watermark := setMap(make(map[string]interface{}))

		watermark.SetString("horizontal_align", w.HorizontalAlign)
		watermark.SetString("horizontal_offset", w.HorizontalOffset)
		watermark.SetString("id", w.Id)
		watermark.SetString("max_height", w.MaxHeight)
		watermark.SetString("max_width", w.MaxWidth)
		watermark.SetString("opacity", w.Opacity)
		watermark.SetString("sizing_policy", w.SizingPolicy)
		watermark.SetString("target", w.Target)
		watermark.SetString("vertical_align", w.VerticalAlign)
		watermark.SetString("vertical_offset", w.VerticalOffset)

		watermarkSet = append(watermarkSet, watermark.Map())
	}

	return watermarkSet
}

func resourceAwsElasticTranscoderPresetDelete(d *schema.ResourceData, meta interface{}) error {
	elastictranscoderconn := meta.(*AWSClient).elastictranscoderconn

	log.Printf("[DEBUG] Elastic Transcoder Delete Preset: %s", d.Id())
	_, err := elastictranscoderconn.DeletePreset(&elastictranscoder.DeletePresetInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting Elastic Transcoder Preset: %s", err)
	}

	return nil
}
